package check

import (
	"getbind2api/common/config"
	logger "getbind2api/common/loggger"
)

func CheckEnvVariable() {
	logger.SysLog("environment variable checking...")

	if config.GBCookie == "" {
		logger.FatalLog("环境变量 USER_ID 未设置")
	}

	logger.SysLog("environment variable check passed.")
}
