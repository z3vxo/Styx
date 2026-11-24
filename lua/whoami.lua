return function(reader)
    local descriptions = {
        SeAssignPrimaryTokenPrivilege = "Replace a process level token",
        SeAuditPrivilege = "Generate security audits",
        SeBackupPrivilege = "Back up files and directories",
        SeChangeNotifyPrivilege = "Bypass traverse checking",
        SeCreateGlobalPrivilege = "Create global objects",
        SeCreatePagefilePrivilege = "Create a pagefile",
        SeCreatePermanentPrivilege = "Create permanent shared objects",
        SeCreateSymbolicLinkPrivilege = "Create symbolic links",
        SeCreateTokenPrivilege = "Create a token object",
        SeDebugPrivilege = "Debug programs",
        SeDelegateSessionUserImpersonatePrivilege = "Obtain an impersonation token for another user in the same session",
        SeEnableDelegationPrivilege = "Enable computer and user accounts to be trusted for delegation",
        SeImpersonatePrivilege = "Impersonate a client after authentication",
        SeIncreaseBasePriorityPrivilege = "Increase scheduling priority",
        SeIncreaseQuotaPrivilege = "Adjust memory quotas for a process",
        SeIncreaseWorkingSetPrivilege = "Increase a process working set",
        SeLoadDriverPrivilege = "Load and unload device drivers",
        SeLockMemoryPrivilege = "Lock pages in memory",
        SeMachineAccountPrivilege = "Add workstations to domain",
        SeManageVolumePrivilege = "Perform volume maintenance tasks",
        SeProfileSingleProcessPrivilege = "Profile single process",
        SeRelabelPrivilege = "Modify an object label",
        SeRemoteShutdownPrivilege = "Force shutdown from a remote system",
        SeRestorePrivilege = "Restore files and directories",
        SeSecurityPrivilege = "Manage auditing and security log",
        SeShutdownPrivilege = "Shut down the system",
        SeSyncAgentPrivilege = "Synchronize directory service data",
        SeSystemEnvironmentPrivilege = "Modify firmware environment values",
        SeSystemProfilePrivilege = "Profile system performance",
        SeSystemtimePrivilege = "Change the system time",
        SeTakeOwnershipPrivilege = "Take ownership of files or other objects",
        SeTcbPrivilege = "Act as part of the operating system",
        SeTimeZonePrivilege = "Change the time zone",
        SeTrustedCredManAccessPrivilege = "Access Credential Manager as a trusted caller",
        SeUndockPrivilege = "Remove computer from docking station",
        SeUnsolicitedInputPrivilege = "Read unsolicited input from a terminal device",
    }

    local states = {
        [0] = "Removed",
        [1] = "Enabled",
        [2] = "Enabled",
        [3] = "Disabled",
    }

    local nameLen = reader:read4()
    local name = reader:readBytes(nameLen)
    local domainLen = reader:read4()
    local domain = reader:readBytes(domainLen)
    local sidLen = reader:read4()
    local sid = reader:readBytes(sidLen)

    local privileges = {}
    local nameWidth = 29
    local descWidth = 36

    while true do
        local privLen = reader:read4()
        if privLen == 0xFFFFFFFF then
            break
        end

        local privName = reader:readBytes(privLen)
        local status = reader:read4()
        local description = descriptions[privName] or ""

        if #privName > nameWidth then nameWidth = #privName end
        if #description > descWidth then descWidth = #description end

        privileges[#privileges + 1] = {
            name = privName,
            description = description,
            state = states[status] or "Unknown",
        }
    end

    local fullUser = name
    if #domain > 0 then
        fullUser = domain .. "\\" .. name
    end

    local userWidth = math.max(22, #fullUser)
    local sidWidth = math.max(46, #sid)
    local stateWidth = 8

    local function pad(value, width)
        return value .. string.rep(" ", width - #value)
    end

    local lines = {
        "USER INFORMATION",
        "----------------",
        "",
        pad("User Name", userWidth) .. " SID",
        string.rep("=", userWidth) .. " " .. string.rep("=", sidWidth),
        pad(fullUser, userWidth) .. " " .. sid,
        "",
        "PRIVILEGES INFORMATION",
        "----------------------",
        "",
        pad("Privilege Name", nameWidth) .. " " .. pad("Description", descWidth) .. " State",
        string.rep("=", nameWidth) .. " " .. string.rep("=", descWidth) .. " " .. string.rep("=", stateWidth),
    }

    for _, privilege in ipairs(privileges) do
        lines[#lines + 1] = pad(privilege.name, nameWidth)
            .. " " .. pad(privilege.description, descWidth)
            .. " " .. privilege.state
    end

    return table.concat(lines, "\n")
end
