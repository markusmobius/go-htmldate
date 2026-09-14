package htmldate

import "syscall"

func init() {
	var information syscall.Timezoneinformation
	if _, err := syscall.GetTimeZoneInformation(&information); err != nil {
		return
	}
	local := &localDateutilEnvironment.timezone
	local.StandardName = syscall.UTF16ToString(information.StandardName[:])
	local.DaylightName = syscall.UTF16ToString(information.DaylightName[:])
	local.StandardOffset = -int(information.Bias+information.StandardBias) * 60
	local.DaylightOffset = local.StandardOffset
	if information.DaylightDate.Month != 0 {
		local.DaylightOffset = -int(information.Bias+information.DaylightBias) * 60
	}
}
