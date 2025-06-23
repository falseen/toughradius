package app

const (
	ConfigSystemTitle         = "SystemTitle"
	ConfigSystemTheme         = "SystemTheme"
	ConfigSystemLoginRemark   = "SystemLoginRemark"
	ConfigSystemLoginSubtitle = "SystemLoginSubtitle"

	ConfigRadiusIgnorePwd             = "RadiusIgnorePwd"
	ConfigRadiusAccountingHistoryDays = "AccountingHistoryDays"
	ConfigRadiusAcctInterimInterval   = "AcctInterimInterval"
	ConfigRadiusEapMethod             = "RadiusEapMethod"
	ConfigRadiusOnlineExpireSeconds   = "OnlineExpireSeconds"

	ConfigTR069AccessAddress           = "TR069AccessAddress"
	ConfigTR069AccessPassword          = "TR069AccessPassword"
	ConfigCpeConnectionRequestPassword = "CpeConnectionRequestPassword"
	ConfigCpeAutoRegister              = "CpeAutoRegister"
)

const (
	RadiusVendorMikrotik = "14988"
	RadiusVendorIkuai    = "10055"
	RadiusVendorHuawei   = "2011"
	RadiusVendorZte      = "3902"
	RadiusVendorH3c      = "25506"
	RadiusVendorRadback  = "2352"
	RadiusVendorCisco    = "9"
	RadiusVendorStandard = "0"
)

var ConfigConstants = []string{
	ConfigSystemTitle,
	ConfigSystemTheme,
	ConfigSystemLoginRemark,
	ConfigSystemLoginSubtitle,
	ConfigTR069AccessAddress,
	ConfigTR069AccessPassword,
	ConfigCpeConnectionRequestPassword,
	ConfigCpeAutoRegister,
	ConfigRadiusIgnorePwd,
	ConfigRadiusAccountingHistoryDays,
	ConfigRadiusAcctInterimInterval,
	ConfigRadiusEapMethod,
	ConfigRadiusOnlineExpireSeconds,
}
