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
}
