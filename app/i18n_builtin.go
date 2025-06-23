package app

// initBuiltinTrans 写入内置中文翻译，用于首次启动。
// 只在 Application.Init 中调用一次即可，多次写入同一 key 不会有副作用。
func (a *Application) initBuiltinTrans() {
	// Radius user duplicate messages
	a.TranslateUpdate(ZhCN, "radius", "Username %s already exists", "用户名 %s 已存在")
	a.TranslateUpdate(ZhCN, "radius", "IP address %s already exists in the current profile", "IP 地址 %s 已在当前策略中存在")
	a.TranslateUpdate(ZhCN, "radius", "row %d username %s already exists", "第 %d 行 用户名 %s 已存在")
	a.TranslateUpdate(ZhCN, "radius", "row %d username %s duplicate within import file", "第 %d 行 用户名 %s 在导入文件中重复")
	a.TranslateUpdate(ZhCN, "radius", "row %d customer or profile is empty", "第 %d 行 客户或策略为空")
	a.TranslateUpdate(ZhCN, "radius", "row %d  Profile：%s does not exist", "第 %d 行 策略：%s 不存在")
	a.TranslateUpdate(ZhCN, "radius", "row %d Expiration time format error：%s", "第 %d 行 过期时间格式错误：%s")
	a.TranslateUpdate(ZhCN, "radius", "row %d ip %s already exists in profile %s", "第 %d 行 IP %s 在策略 %s 中已存在")
	a.TranslateUpdate(ZhCN, "radius", "row %d ip %s duplicate within import file for profile %s", "第 %d 行 IP %s 在导入文件中于策略 %s 重复")
	a.TranslateUpdate(ZhCN, "radius", "row %d ip %s already exists in DB", "第 %d 行 IP %s 已存在")
	a.TranslateUpdate(ZhCN, "radius", "row %d ip %s duplicate within import file", "第 %d 行 IP %s 在导入文件中重复")
	a.TranslateUpdate(ZhCN, "radius", "AccessType", "账号类型")
	a.TranslateUpdate(ZhCN, "radius", "Auto", "自动")
	a.TranslateUpdate(ZhCN, "radius", "PPPoE", "PPPoE")
	a.TranslateUpdate(ZhCN, "radius", "VPN", "VPN")
	a.TranslateUpdate(ZhCN, "settings", "Online record expire seconds", "在线记录过期秒数")
	a.TranslateUpdate(ZhCN, "settings", "Delete online record if last update exceeds this value (seconds)", "当在线记录最后更新时间超过此秒数即判定离线")
}
