package radius

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/v8/app"
	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/web"
	"github.com/talkincode/toughradius/v8/models"
	"github.com/talkincode/toughradius/v8/webserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InitUserRouter Radius User additions, deletions, modification and query
func InitUserRouter() {

	webserver.GET("/admin/radius/users", func(c echo.Context) error {
		return c.Render(http.StatusOK, "radius_users", map[string]interface{}{})
	})

	webserver.GET("/admin/radius/users/options", func(c echo.Context) error {
		var data []models.RadiusUser
		query := app.GDB().Model(&models.RadiusUser{})
		cid := c.QueryParam("node_id")
		if cid != "" {
			query = query.Where("node_id = ?", cid)
		}
		ids := c.QueryParam("ids")
		if ids != "" {
			query = query.Where("id in (?)", strings.Split(ids, ","))
		}
		common.Must(query.Find(&data).Error)
		var options = make([]web.JsonOptions, 0)
		for _, d := range data {
			options = append(options, web.JsonOptions{
				Id:    cast.ToString(d.ID),
				Value: d.Username,
			})
		}
		return c.JSON(http.StatusOK, options)
	})

	webserver.GET("/admin/radius/users/query", func(c echo.Context) error {
		var count, start, expireDays int
		web.NewParamReader(c).
			ReadInt(&start, "start", 0).
			ReadInt(&count, "count", 40).
			ReadInt(&expireDays, "expire_days", 0)
		var data []models.RadiusUser
		getQuery := func() *gorm.DB {
			// query := app.GDB().Model(&models.RadiusUser{})
			query := app.GDB().Model(&models.RadiusUser{}).Select("radius_user.*, coalesce(ro.count, 0) as online_count").
				Joins("left join (select username, count(1) as count from radius_online  group by username) ro on radius_user.username = ro.username")
			if len(web.ParseSortMap(c)) == 0 {
				query = query.Order("radius_user.username asc")
			} else {
				mobj := models.RadiusUser{}
				for name, stype := range web.ParseSortMap(c) {
					if common.GetFieldType(mobj, name) == "string" {
						query = query.Order(fmt.Sprintf("convert_to(radius_user.%s,'UTF8')  %s ", name, stype))
					} else {
						if name == "online_count" {
							query = query.Order(fmt.Sprintf("%s %s ", name, stype))
						} else {
							query = query.Order(fmt.Sprintf("radius_user.%s %s ", name, stype))
						}
					}
				}
			}

			for name, value := range web.ParseEqualMap(c) {
				query = query.Where(fmt.Sprintf("radius_user.%s = ?", name), value)
			}

			for name, value := range web.ParseFilterMap(c) {
				if common.InSlice(name, []string{"profile_id", "pnode_id"}) {
					query = query.Where(fmt.Sprintf("radius_user.%s = ?", name), value)
				} else {
					query = query.Where(fmt.Sprintf("radius_user.%s like ?", name), "%"+value+"%")
				}
			}

			if expireDays > 1 {
				query = query.Where("radius_user.expire_time <=  ?", time.Now().Add(time.Hour*24*time.Duration(expireDays)))
			}

			keyword := c.QueryParam("keyword")
			if keyword != "" {
				query = query.Where("radius_user.username like ?", "%"+keyword+"%").
					Or("radius_user.remark like ?", "%"+keyword+"%").
					Or("radius_user.realname like ?", "%"+keyword+"%").
					Or("radius_user.mobile like ?", "%"+keyword+"%")
			}
			return query
		}
		var total int64
		common.Must(getQuery().Count(&total).Error)

		query := getQuery().Offset(start).Limit(count)
		if query.Find(&data).Error != nil {
			return c.JSON(http.StatusOK, common.EmptyList)
		}
		return c.JSON(http.StatusOK, &web.PageResult{TotalCount: total, Pos: int64(start), Data: data})
	})

	webserver.GET("/admin/radius/users/get", func(c echo.Context) error {
		var id string
		web.NewParamReader(c).
			ReadRequiedString(&id, "id")
		var data models.RadiusUser
		common.Must(app.GDB().Where("id=?", id).First(&data).Error)
		return c.JSON(http.StatusOK, data)
	})

	webserver.POST("/admin/radius/users/add", func(c echo.Context) error {
		form := new(models.RadiusUser)
		common.Must(c.Bind(form))
		form.ID = common.UUIDint64()
		form.CreatedAt = time.Now()
		form.UpdatedAt = time.Now()
		timestr := c.FormValue("expire_time")[:10] + " 23:59:59"
		form.ExpireTime, _ = time.Parse("2006-01-02 15:04:05", timestr)
		common.CheckEmpty("username", form.Username)
		common.CheckEmpty("password", form.Password)

		var count int64 = 0
		app.GDB().Model(models.RadiusUser{}).Where("username=?", form.Username).Count(&count)
		if count > 0 {
			return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "Username %s already exists"), form.Username)))
		}

		// Check duplicate IP in the same profile when IP is specified
		if strings.TrimSpace(form.IpAddr) != "" {
			var ipcnt int64
			app.GDB().Model(models.RadiusUser{}).Where("profile_id = ? and ip_addr = ?", form.ProfileId, form.IpAddr).Count(&ipcnt)
			if ipcnt > 0 {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "IP address %s already exists in the current profile"), form.IpAddr)))
			}
		}

		var profile models.RadiusProfile
		common.Must(app.GDB().Where("id=?", form.ProfileId).First(&profile).Error)

		form.ActiveNum = profile.ActiveNum
		form.UpRate = profile.UpRate
		form.DownRate = profile.DownRate
		form.AddrPool = profile.AddrPool
		// 接入类型继承策略，除非前端显式指定
		if strings.TrimSpace(form.AccessType) == "" {
			form.AccessType = profile.AccessType
		}

		common.Must(app.GDB().Create(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Create RADIUS user：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/update", func(c echo.Context) error {
		form := new(models.RadiusUser)
		common.Must(c.Bind(form))
		timestr := c.FormValue("expire_time")[:10] + " 23:59:59"
		form.ExpireTime, _ = time.Parse("2006-01-02 15:04:05", timestr)
		common.CheckEmpty("username", form.Username)
		common.CheckEmpty("password", form.Password)
		var profile models.RadiusProfile
		common.Must(app.GDB().Where("id=?", form.ProfileId).First(&profile).Error)
		form.ActiveNum = profile.ActiveNum
		form.UpRate = profile.UpRate
		form.DownRate = profile.DownRate
		form.AddrPool = profile.AddrPool
		if strings.TrimSpace(form.AccessType) == "" {
			form.AccessType = profile.AccessType
		}

		common.Must(app.GDB().Save(form).Error)
		webserver.PubOpLog(c, fmt.Sprintf("Update RADIUS users：%v", form))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/batchupdate", func(c echo.Context) error {
		profileId := c.FormValue("profile_id")
		userIds := c.FormValue("user_ids")
		expireDay := c.FormValue("expire_time")
		status := c.FormValue("status")
		if userIds == "" {
			return c.JSON(200, web.RestError("No user account selected"))
		}
		expire, err := time.Parse("2006-01-02", expireDay[:10])
		if err != nil {
			return c.JSON(200, web.RestError("Wrong time format"))
		}
		var profileNone = false
		var profile models.RadiusProfile
		err = app.GDB().Where("id=?", profileId).First(&profile).Error
		if err != nil {
			profileNone = true
		}

		var succ, errs int
		for _, uid := range strings.Split(userIds, ",") {
			var user models.RadiusUser
			err = app.GDB().Where("id=?", uid).First(&user).Error
			if err != nil {
				errs++
				continue
			}
			var data = map[string]interface{}{
				"expire_time": expire,
				"updated_at":  time.Now(),
			}

			if !profileNone {
				data["profile_id"] = profileId
				data["active_num"] = profile.ActiveNum
				data["up_rate"] = profile.UpRate
				data["down_rate"] = profile.DownRate
				data["addr_pool"] = profile.AddrPool
				data["access_type"] = profile.AccessType
			}
			if common.InSlice(status, []string{"enabled", "disabled"}) {
				data["status"] = status
			}

			r := app.GDB().Debug().Model(&models.RadiusUser{}).Where("id=?", strings.TrimSpace(uid)).Updates(&data)
			if r.Error != nil {
				errs += 1
			} else {
				if r.RowsAffected > 0 {
					succ += 1
				}
			}
		}
		webserver.PubOpLog(c, fmt.Sprintf("Update RADIUS users in batches：succ=%d, errs=%d", succ, errs))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.GET("/admin/radius/users/delete", func(c echo.Context) error {
		ids := c.QueryParam("ids")
		common.Must(app.GDB().Delete(models.RadiusUser{}, strings.Split(ids, ",")).Error)
		webserver.PubOpLog(c, fmt.Sprintf("delete RADIUS user：%s", ids))
		return c.JSON(http.StatusOK, web.RestSucc("success"))
	})

	webserver.POST("/admin/radius/users/import", func(c echo.Context) error {
		// 解析上传文件，得到 []map[string]interface{}
		rawRows, err := webserver.ImportData(c, "radius_users")
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError(err.Error()))
		}

		var users []models.RadiusUser
		for _, row := range rawRows {
			// 将表头统一转为小写，解决 Excel 首字母大写/全大写等问题
			lc := make(map[string]interface{}, len(row))
			for k, v := range row {
				lc[strings.ToLower(strings.TrimSpace(k))] = v
			}

			// 如果 id 缺失，用雪花算法补一个，确保主键存在且类型正确
			var uid int64
			if vid, ok := lc["id"]; ok {
				uid = cast.ToInt64(vid)
			} else {
				uid = common.UUIDint64()
			}

			username := strings.ToLower(strings.TrimSpace(cast.ToString(lc["username"])))
			if username == "" {
				// 用户名必填，跳过空行
				continue
			}

			// 基本字段直接类型转换即可
			user := models.RadiusUser{
				ID:         uid,
				NodeId:     cast.ToInt64(lc["nodeid"]),
				ProfileId:  cast.ToInt64(lc["profileid"]),
				Realname:   cast.ToString(lc["realname"]),
				Mobile:     cast.ToString(lc["mobile"]),
				Username:   username,
				Password:   cast.ToString(lc["password"]),
				AddrPool:   cast.ToString(lc["addrpool"]),
				ActiveNum:  cast.ToInt(lc["activenum"]),
				UpRate:     cast.ToInt(lc["uprate"]),
				DownRate:   cast.ToInt(lc["downrate"]),
				Vlanid1:    cast.ToInt(lc["vlanid1"]),
				Vlanid2:    cast.ToInt(lc["vlanid2"]),
				IpAddr:     cast.ToString(lc["ipaddr"]),
				MacAddr:    cast.ToString(lc["macaddr"]),
				BindVlan:   cast.ToInt(lc["bindvlan"]),
				BindMac:    cast.ToInt(lc["bindmac"]),
				Status:     cast.ToString(lc["status"]),
				Remark:     cast.ToString(lc["remark"]),
				AccessType: cast.ToString(lc["accesstype"]),
			}

			// 解析过期时间，允许空
			if v := cast.ToString(lc["expiretime"]); strings.TrimSpace(v) != "" {
				if t, err := time.Parse("2006-01-02", v[:10]); err == nil {
					user.ExpireTime = t
				}
			}

			// 如果 AccessType 为空，让它继承 profile 策略时再填充

			users = append(users, user)
		}

		if len(users) == 0 {
			return c.JSON(http.StatusOK, web.RestError("no valid rows"))
		}

		// 使用 OnConflict(username) 做 upsert，保持与旧逻辑一致
		if err := app.GDB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "username"}},
			UpdateAll: true,
		}).Create(&users).Error; err != nil {
			return c.JSON(http.StatusOK, web.RestError(err.Error()))
		}

		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.POST("/admin/radius/users/importadd", func(c echo.Context) error {
		datas, err := webserver.ImportData(c, "radius_users")
		common.Must(err)
		var unames []string
		var unames2 []string // track duplicate usernames within current import
		var ips2 []string    // track duplicate ip addresses within current import (profile_id:ip)
		app.GDB().Model(models.RadiusUser{}).Pluck("username", &unames)
		var users []models.RadiusUser
		for row, data := range datas {
			// 将列名统一转为小写，兼容 Excel 表头大小写差异
			lc := make(map[string]interface{}, len(data))
			for k, v := range data {
				lc[strings.ToLower(strings.TrimSpace(k))] = v
			}

			username := strings.ToLower(strings.TrimSpace(cast.ToString(lc["username"])))
			if common.InSlice(username, unames) {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d username %s already exists"), row+1, username)))
			}
			if common.InSlice(username, unames2) {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d username %s duplicate within import file"), row+1, username)))
			}
			unames2 = append(unames2, username)
			password := cast.ToString(lc["password"])
			expire := cast.ToString(lc["expiretime"])
			mobile := cast.ToString(lc["mobile"])
			remark := cast.ToString(lc["remark"])
			realname := strings.TrimSpace(cast.ToString(lc["realname"]))
			if realname == "" {
				realname = username
			}
			common.CheckEmpty("username", username)
			common.CheckEmpty("password", password)
			common.CheckEmpty("expire_time", expire)
			profileName := cast.ToString(lc["profile"])
			if profileName == "" {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d customer or profile is empty"), row+1)))
			}
			var profile models.RadiusProfile
			err := app.GDB().Where("name=?", profileName).First(&profile).Error
			if err != nil {
				return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d  Profile：%s does not exist"), row+1, profileName)))
			}
			expireTime, err := time.Parse("2006-01-02", expire)
			if err != nil {
				expireTime, err = time.Parse("2006/01/02", expire)
				if err != nil {
					return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d Expiration time format error：%s"), row+1, expire)))
				}
			}
			ipAddr := strings.TrimSpace(cast.ToString(lc["ipaddr"]))
			// Validate duplicate IP in DB for this profile
			if ipAddr != "" {
				var ipCnt int64
				app.GDB().Model(models.RadiusUser{}).Where("profile_id = ? and ip_addr = ?", profile.ID, ipAddr).Count(&ipCnt)
				if ipCnt > 0 {
					return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d ip %s already exists in profile %s"), row+1, ipAddr, profile.Name)))
				}
				// Validate duplicate IP within current import batch
				var profileIpKey = fmt.Sprintf("%d:%s", profile.ID, ipAddr)
				if common.InSlice(profileIpKey, ips2) {
					return c.JSON(http.StatusOK, web.RestError(fmt.Sprintf(app.Trans("radius", "row %d ip %s duplicate within import file for profile %s"), row+1, ipAddr, profile.Name)))
				}
				ips2 = append(ips2, profileIpKey)
			}
			user := models.RadiusUser{
				ID:          common.UUIDint64(),
				NodeId:      profile.NodeId,
				ProfileId:   profile.ID,
				Realname:    realname,
				Mobile:      mobile,
				Username:    username,
				Password:    password,
				AddrPool:    profile.AddrPool,
				ActiveNum:   profile.ActiveNum,
				UpRate:      profile.UpRate,
				DownRate:    profile.DownRate,
				AccessType:  profile.AccessType,
				IpAddr:      ipAddr,
				ExpireTime:  expireTime,
				Status:      "enabled",
				Remark:      remark,
				OnlineCount: 0,
				LastOnline:  time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			users = append(users, user)
		}
		err = app.GDB().Save(&users).Error
		if err != nil {
			return c.JSON(http.StatusOK, web.RestError(err.Error()))
		}
		return c.JSON(http.StatusOK, web.RestSucc("Success"))
	})

	webserver.GET("/admin/radius/users/export", func(c echo.Context) error {
		var data []models.RadiusUser
		common.Must(app.GDB().Find(&data).Error)
		return webserver.ExportCsv(c, data, "radius_users")
	})
}
