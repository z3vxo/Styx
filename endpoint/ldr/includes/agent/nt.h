#pragma once
#include <windows.h>


#define NTAPI_SUCCESS(Status) (((NTSTATUS)(Status)) >= 0)

typedef enum _PROCESSINFOCLASS
{
    ProcessBasicInformation,                        // q: PROCESS_BASIC_INFORMATION, PROCESS_EXTENDED_BASIC_INFORMATION
    ProcessQuotaLimits,                             // qs: QUOTA_LIMITS, QUOTA_LIMITS_EX
    ProcessIoCounters,                              // q: IO_COUNTERS
    ProcessVmCounters,                              // q: VM_COUNTERS, VM_COUNTERS_EX, VM_COUNTERS_EX2
    ProcessTimes,                                   // q: KERNEL_USER_TIMES // since VISTA
    ProcessBasePriority,                            // s: KPRIORITY (raising priority requires SeIncreaseBasePriorityPrivilege)
    ProcessRaisePriority,                           // s: PROCESS_RAISE_PRIORITY
    ProcessDebugPort,                               // q: HANDLE
    ProcessExceptionPort,                           // s: HANDLE or PROCESS_EXCEPTION_PORT (requires SeTcbPrivilege)
    ProcessAccessToken,                             // s: PROCESS_ACCESS_TOKEN
    ProcessLdtInformation,                          // qs: PROCESS_LDT_INFORMATION // 10
    ProcessLdtSize,                                 // s: PROCESS_LDT_SIZE
    ProcessDefaultHardErrorMode,                    // qs: PROCESS_DEFAULT_HARD_ERROR_MODE
    ProcessIoPortHandlers,                          // s: PROCESS_IO_PORT_HANDLER_INFORMATION // (kernel-mode only)
    ProcessPooledUsageAndLimits,                    // q: POOLED_USAGE_AND_LIMITS
    ProcessWorkingSetWatch,                         // q: PROCESS_WS_WATCH_INFORMATION[]; s: void
    ProcessUserModeIOPL,                            // s: PROCESS_USER_MODE_IOPL (requires SeTcbPrivilege)
    ProcessEnableAlignmentFaultFixup,               // s: BOOLEAN
    ProcessPriorityClass,                           // qs: PROCESS_PRIORITY_CLASS
    ProcessWx86Information,                         // qs: ULONG (requires SeTcbPrivilege) (VdmAllowed)
    ProcessHandleCount,                             // q: ULONG, PROCESS_HANDLE_INFORMATION // 20
    ProcessAffinityMask,                            // qs: KAFFINITY, GROUP_AFFINITY
    ProcessPriorityBoost,                           // qs: PROCESS_PRIORITY_BOOST
    ProcessDeviceMap,                               // q: PROCESS_DEVICEMAP_INFORMATION.Query, PROCESS_DEVICEMAP_INFORMATION_EX; s: PROCESS_DEVICEMAP_INFORMATION.Set
    ProcessSessionInformation,                      // qs: PROCESS_SESSION_INFORMATION (s: requires SeTcbPrivilege)
    ProcessForegroundInformation,                   // s: PROCESS_FOREGROUND_BACKGROUND
    ProcessWow64Information,                        // q: ULONG_PTR
    ProcessImageFileName,                           // q: UNICODE_STRING
    ProcessLUIDDeviceMapsEnabled,                   // q: PROCESS_LUID_DEVICE_MAPS_ENABLED
    ProcessBreakOnTermination,                      // qs: ULONG (s: requires SeDebugPrivilege)
    ProcessDebugObjectHandle,                       // q: HANDLE // 30
    ProcessDebugFlags,                              // qs: PROCESS_DEBUG_FLAGS
    ProcessHandleTracing,                           // q: PROCESS_HANDLE_TRACING_QUERY; s: PROCESS_HANDLE_TRACING_ENABLE[_EX] or void to disable
    ProcessIoPriority,                              // qs: IO_PRIORITY_HINT; s: PROCESS_IO_PRIORITY_INFORMATION (IoPriorityHigh requires SeIncreaseBasePriorityPrivilege)
    ProcessExecuteFlags,                            // qs: PROCESS_EXECUTE_FLAGS (s: NtCurrentProcess only)
    ProcessTlsInformation,                          // s: PROCESS_TLS_INFORMATION (NtCurrentProcess, user-mode only) // ProcessResourceManagement
    ProcessCookie,                                  // q: ULONG
    ProcessImageInformation,                        // q: SECTION_IMAGE_INFORMATION
    ProcessCycleTime,                               // q: PROCESS_CYCLE_TIME_INFORMATION // since VISTA
    ProcessPagePriority,                            // qs: PAGE_PRIORITY_INFORMATION
    ProcessInstrumentationCallback,                 // s: PVOID or PROCESS_INSTRUMENTATION_CALLBACK_INFORMATION // 40
    ProcessThreadStackAllocation,                   // s: PROCESS_STACK_ALLOCATION_INFORMATION, PROCESS_STACK_ALLOCATION_INFORMATION_EX (NtCurrentProcess only)
    ProcessWorkingSetWatchEx,                       // q: PROCESS_WS_WATCH_INFORMATION_EX[]; s: void
    ProcessImageFileNameWin32,                      // q: UNICODE_STRING
    ProcessImageFileMapping,                        // q: HANDLE (input)
    ProcessAffinityUpdateMode,                      // qs: PROCESS_AFFINITY_UPDATE_MODE (s: NtCurrentProcess only)
    ProcessMemoryAllocationMode,                    // qs: PROCESS_MEMORY_ALLOCATION_MODE
    ProcessGroupInformation,                        // q: USHORT[] (group numbers; ReturnLength required)
    ProcessTokenVirtualizationEnabled,              // s: ULONG
    ProcessConsoleHostProcess,                      // qs: PROCESS_CONSOLE_HOST_PROCESS_INFORMATION (s: NtCurrentProcess only)
    ProcessWindowInformation,                       // q: PROCESS_WINDOW_INFORMATION // 50
    ProcessHandleInformation,                       // q: PROCESS_HANDLE_SNAPSHOT_INFORMATION // since WIN8
    ProcessMitigationPolicy,                        // qs: PROCESS_MITIGATION_POLICY_INFORMATION
    ProcessDynamicFunctionTableInformation,         // s: PROCESS_DYNAMIC_FUNCTION_TABLE_INFORMATION (NtCurrentProcess only)
    ProcessHandleCheckingMode,                      // qs: PROCESS_HANDLE_CHECKING_MODE; s: 0 disables, 1 enables, other values invalid
    ProcessKeepAliveCount,                          // q: PROCESS_KEEPALIVE_COUNT_INFORMATION
    ProcessRevokeFileHandles,                       // s: PROCESS_REVOKE_FILE_HANDLES_INFORMATION
    ProcessWorkingSetControl,                       // s: PROCESS_WORKING_SET_CONTROL
    ProcessHandleTable,                             // q: ULONG[] // since WINBLUE
    ProcessCheckStackExtentsMode,                   // qs: ULONG (s: remote process only, requires SeDebugPrivilege) // KPROCESS->CheckStackExtents (CFG)
    ProcessCommandLineInformation,                  // q: UNICODE_STRING // 60
    ProcessProtectionInformation,                   // q: PS_PROTECTION
    ProcessMemoryExhaustion,                        // s: PROCESS_MEMORY_EXHAUSTION_INFO // since THRESHOLD
    ProcessFaultInformation,                        // s: PROCESS_FAULT_INFORMATION
    ProcessTelemetryIdInformation,                  // q: PROCESS_TELEMETRY_ID_INFORMATION
    ProcessCommitReleaseInformation,                // qs: PROCESS_COMMIT_RELEASE_INFORMATION
    ProcessDefaultCpuSetsInformation,               // qs: ULONG64[] (CPU-set masks; length 0..256 bytes, multiple of 8) // ProcessReserved1Information
    ProcessAllowedCpuSetsInformation,               // qs: ULONG64[] (CPU-set masks; length 0..256 bytes, multiple of 8; s: CPU-set resource-manager access check) // ProcessReserved2Information
    ProcessSubsystemProcess,                        // s: void // EPROCESS->SubsystemProcess
    ProcessJobMemoryInformation,                    // q: PROCESS_JOB_MEMORY_INFO
    ProcessInPrivate,                               // q: BOOLEAN; s: void // ETW // since THRESHOLD2 // 70
    ProcessRaiseUMExceptionOnInvalidHandleClose,    // qs: PROCESS_RAISE_UM_EXCEPTION_ON_INVALID_HANDLE_CLOSE; s: 0 disables, 1 enables, other values leave unchanged
    ProcessIumChallengeResponse,                    // qs: PROCESS_IUM_CHALLENGE_RESPONSE
    ProcessChildProcessInformation,                 // q: PROCESS_CHILD_PROCESS_INFORMATION
    ProcessHighGraphicsPriorityInformation,         // qs: BOOLEAN (s: requires SeTcbPrivilege)
    ProcessSubsystemInformation,                    // q: SUBSYSTEM_INFORMATION_TYPE // since REDSTONE2
    ProcessEnergyValues,                            // q: PROCESS_ENERGY_VALUES, PROCESS_EXTENDED_ENERGY_VALUES, PROCESS_EXTENDED_ENERGY_VALUES_V1
    ProcessPowerThrottlingState,                    // qs: POWER_THROTTLING_PROCESS_STATE
    ProcessActivityThrottlePolicy,                  // qs: Obsolete // PROCESS_ACTIVITY_THROTTLE_POLICY // ProcessReserved3Information
    ProcessWin32kSyscallFilterInformation,          // q: WIN32K_SYSCALL_FILTER
    ProcessDisableSystemAllowedCpuSets,             // s: BOOLEAN (CPU-set resource-manager access check) // 80
    ProcessWakeInformation,                         // q: PROCESS_WAKE_INFORMATION // (kernel-mode only)
    ProcessEnergyTrackingState,                     // qs: PROCESS_ENERGY_TRACKING_STATE
    ProcessManageWritesToExecutableMemory,          // s: MANAGE_WRITES_TO_EXECUTABLE_MEMORY // since REDSTONE3
    ProcessCaptureTrustletLiveDump,                 // q: ULONG
    ProcessTelemetryCoverage,                       // q: TELEMETRY_COVERAGE_HEADER; s: TELEMETRY_COVERAGE_POINT
    ProcessEnclaveInformation,                      // qs: Obsolete
    ProcessEnableReadWriteVmLogging,                // qs: PROCESS_READWRITEVM_LOGGING_INFORMATION (s: requires SeDebugPrivilege or SeTcbPrivilege, and security-logger access)
    ProcessUptimeInformation,                       // q: PROCESS_UPTIME_INFORMATION
    ProcessImageSection,                            // q: HANDLE (NtCurrentProcess, not attached to another process)
    ProcessDebugAuthInformation,                    // s: PROCESS_DEBUG_AUTH_INFORMATION // CiTool.exe -- device-id // PplDebugAuthorization // since RS4 // 90
    ProcessSystemResourceManagement,                // s: PROCESS_SYSTEM_RESOURCE_MANAGEMENT (requires SeTcbPrivilege)
    ProcessSequenceNumber,                          // q: ULONGLONG
    ProcessLoaderDetour,                            // qs: Obsolete // since RS5
    ProcessSecurityDomainInformation,               // q: PROCESS_SECURITY_DOMAIN_INFORMATION
    ProcessCombineSecurityDomainsInformation,       // s: PROCESS_COMBINE_SECURITY_DOMAINS_INFORMATION
    ProcessEnableLogging,                           // qs: PROCESS_LOGGING_INFORMATION (s: requires SeDebugPrivilege or SeTcbPrivilege, and security-logger access)
    ProcessLeapSecondInformation,                   // qs: PROCESS_LEAP_SECOND_INFORMATION (NtCurrentProcess only)
    ProcessFiberShadowStackAllocation,              // s: PROCESS_FIBER_SHADOW_STACK_ALLOCATION_INFORMATION (NtCurrentProcess, user-mode only) // since 19H1
    ProcessFreeFiberShadowStackAllocation,          // s: PROCESS_FREE_FIBER_SHADOW_STACK_ALLOCATION_INFORMATION (NtCurrentProcess, user-mode only)
    ProcessAltSystemCallInformation,                // s: PROCESS_SYSCALL_PROVIDER_INFORMATION // since 20H1 // 100
    ProcessDynamicEHContinuationTargets,            // s: PROCESS_DYNAMIC_EH_CONTINUATION_TARGETS_INFORMATION
    ProcessDynamicEnforcedCetCompatibleRanges,      // s: PROCESS_DYNAMIC_ENFORCED_ADDRESS_RANGE_INFORMATION // since 20H2
    ProcessCreateStateChange,                       // qs: Obsolete // since WIN11
    ProcessApplyStateChange,                        // qs: Obsolete
    ProcessEnableOptionalXStateFeatures,            // s: ULONG64 (current process, user-mode only) // EnableProcessOptionalXStateFeatures
    ProcessAltPrefetchParam,                        // qs: OVERRIDE_PREFETCH_PARAMETER // App Launch Prefetch (ALPF) // since 22H1
    ProcessAssignCpuPartitions,                     // s: PROCESS_ASSIGN_CPU_PARTITIONS_INFORMATION
    ProcessPriorityClassEx,                         // s: PROCESS_PRIORITY_CLASS_EX
    ProcessMembershipInformation,                   // q: PROCESS_MEMBERSHIP_INFORMATION
    ProcessEffectiveIoPriority,                     // q: IO_PRIORITY_HINT // 110
    ProcessEffectivePagePriority,                   // q: ULONG
    ProcessSchedulerSharedData,                     // s: PROCESS_SCHEDULER_SHARED_DATA_SLOT_INFORMATION (output) // since 24H2
    ProcessSlistRollbackInformation,                // s: void (length 0, NtCurrentProcess only)
    ProcessNetworkIoCounters,                       // q: PROCESS_NETWORK_COUNTERS
    ProcessFindFirstThreadByTebValue,               // q: PROCESS_TEB_VALUE_INFORMATION // NtCurrentProcess
    ProcessEnclaveAddressSpaceRestriction,          // qs: Obsolete // since 25H2
    ProcessAvailableCpus,                           // qs: Obsolete // PROCESS_AVAILABLE_CPUS_INFORMATION
    MaxProcessInfoClass
} PROCESSINFOCLASS;


#define InitializeObjectAttributes( p , n , a, r, s) { \
    (p)->Length = sizeof(OBJECT_ATTRIBUTES);           \
    (p)->RootDirectory = r;                            \
    (p)->Attributes    = a;                            \
    (p)->ObjectName    = n;                            \
    (p)->SecurityDescriptor = s;                       \
    (p)->SecurityQualityOfService = NULL;              \
}
typedef struct _UNICODE_STRING {
    USHORT Length;
    USHORT MaxiumLength;
    PWSTR Buffer;
}UNICODE_STRING, * PUNICODE_STRING;

typedef const UNICODE_STRING* PCUNICODE_STRING;


typedef struct _OBJECT_ATTRIBUTES
{
    ULONG Length;
    HANDLE RootDirectory;
    PCUNICODE_STRING ObjectName;
    ULONG Attributes;
    PSECURITY_DESCRIPTOR SecurityDescriptor;
    PSECURITY_QUALITY_OF_SERVICE SecurityQualityOfService;
} OBJECT_ATTRIBUTES, * POBJECT_ATTRIBUTES;

typedef const OBJECT_ATTRIBUTES* PCOBJECT_ATTRIBUTES;

typedef struct _CLIENT_ID {
    HANDLE UniqueProcess;
    HANDLE UniqueThread;
}CLIENT_ID, * PCLIENT_ID;

typedef enum _SECTION_INHERIT
{
    ViewShare = 1, // The mapped view of the section will be mapped into any child processes created by the process.
    ViewUnmap = 2  // The mapped view of the section will not be mapped into any child processes created by the process.
} SECTION_INHERIT;

extern "C" {
    NTSYSAPI NTSTATUS NTAPI RtlGetVersion(PRTL_OSVERSIONINFOW);
    NTSYSAPI NTSTATUS NTAPI RtlGetNtVersionNumbers(DWORD* major, DWORD* minor, DWORD* build);
    NTSYSAPI NTSTATUS NTAPI RtlRandomEx(PULONG seed);
    NTSYSAPI ULONG    NTAPI RtlNtStatusToDosError(NTSTATUS Status);
    NTSYSAPI NTSTATUS NTAPI NtOpenProcess(
        PHANDLE ProcessHandle,
        ACCESS_MASK DesiredAccess,
        PCOBJECT_ATTRIBUTES ObjectAttributes,
        PCLIENT_ID ClientId
    );
    NTSYSCALLAPI NTSTATUS NTAPI NtOpenProcessToken(
        _In_ HANDLE ProcessHandle,
        _In_ ACCESS_MASK DesiredAccess,
        _Out_ PHANDLE TokenHandle
    );
    NTSYSCALLAPI NTSTATUS NTAPI NtQueryInformationProcess(
        _In_ HANDLE ProcessHandle,
        _In_ PROCESSINFOCLASS ProcessInformationClass,
        _Out_writes_bytes_(ProcessInformationLength) PVOID ProcessInformation,
        _In_ ULONG ProcessInformationLength,
        _Out_opt_ PULONG ReturnLength
    );

    NTSYSCALLAPI NTSTATUS NTAPI NtQueryInformationToken(
        _In_ HANDLE TokenHandle,
        _In_ TOKEN_INFORMATION_CLASS TokenInformationClass,
        _Out_writes_bytes_to_opt_(TokenInformationLength, *ReturnLength) PVOID TokenInformation,
        _In_ ULONG TokenInformationLength,
        _Out_ PULONG ReturnLength
    );

    NTSYSCALLAPI NTSTATUS NTAPI NtCreateSection(
              PHANDLE            SectionHandle,
              ACCESS_MASK        DesiredAccess,
              POBJECT_ATTRIBUTES ObjectAttributes,
              PLARGE_INTEGER     MaximumSize,
              ULONG              SectionPageProtection,
              ULONG              AllocationAttributes,
              HANDLE             FileHandle
    );

    NTSYSCALLAPI NTSTATUS NTAPI NtMapViewOfSection(
               HANDLE          SectionHandle,
               HANDLE          ProcessHandle,
               PVOID           *BaseAddress,
               ULONG_PTR       ZeroBits,
               SIZE_T          CommitSize,
               PLARGE_INTEGER  SectionOffset,
               PSIZE_T         ViewSize,
               SECTION_INHERIT InheritDisposition,
               ULONG           AllocationType,
               ULONG           Win32Protect
);
}


typedef LONG KPRIORITY, * PKPRIORITY;



typedef struct _PEB_LDR_DATA {
    ULONG Len;
    BOOLEAN Intitalised;
    HANDLE SsHandle;
    LIST_ENTRY InLoadOrderModuleList;
    LIST_ENTRY InMemoryOrderModuleList;
    LIST_ENTRY InInitialiszationOrderModuleList;
    PVOID EntryInProgress;
    BOOLEAN ShutDownInProgress;
    HANDLE ShutdownThreadID;
}PEB_LDR_DATA, * PPEB_LDR_DATA;

typedef struct _LDR_DATA_TABLE_ENTRY {
    LIST_ENTRY InLoadOrderLinks;
    LIST_ENTRY InMemoryOrderLinks;
    LIST_ENTRY InInitializationOrderLinks;
    PVOID DllBase;
    PVOID EntryBase;
    ULONG SizeOfImage;
    UNICODE_STRING FullDllName;
    UNICODE_STRING BaseDllName;
}LDR_DATA_TABLE_ENTRY, * PLDR_DATA_TABLE_ENTRY;

typedef struct _PEB {
    BOOLEAN InheritedAddressSpace;
    BOOLEAN ReadImageFileExecOptions;
    BOOLEAN BeingDebugged;
    BOOLEAN BitFields;
    HANDLE Mutant;
    PVOID ImageBase;
    PPEB_LDR_DATA Ldr;
}PEB, * PPEB;


typedef struct _PROCESS_BASIC_INFORMATION
{
    NTSTATUS ExitStatus;                    // The exit status of the process. (GetExitCodeProcess)
    PPEB PebBaseAddress;                    // A pointer to the process environment block (PEB) of the process.
    KAFFINITY AffinityMask;                 // The affinity mask of the process. (GetProcessAffinityMask) (deprecated)
    KPRIORITY BasePriority;                 // The base priority of the process. (GetPriorityClass)
    HANDLE UniqueProcessId;                 // The unique identifier of the process. (GetProcessId)
    HANDLE InheritedFromUniqueProcessId;    // The unique identifier of the parent process.
} PROCESS_BASIC_INFORMATION, * PPROCESS_BASIC_INFORMATION;

typedef struct _ASSEMBLY_STORAGE_MAP_ENTRY
{
    ULONG Flags;
    UNICODE_STRING DosPath;
    HANDLE Handle;
} ASSEMBLY_STORAGE_MAP_ENTRY, * PASSEMBLY_STORAGE_MAP_ENTRY;

typedef struct _ASSEMBLY_STORAGE_MAP
{
    ULONG Flags; // ASSEMBLY_STORAGE_MAP_*
    ULONG AssemblyCount; // includes the reserved roster entry at index zero
    PASSEMBLY_STORAGE_MAP_ENTRY* AssemblyArray; // indexed by assembly roster index; entries may be NULL
} ASSEMBLY_STORAGE_MAP, * PASSEMBLY_STORAGE_MAP;

typedef struct _ACTIVATION_CONTEXT_DATA
{
    ULONG Magic;
    ULONG HeaderSize;
    ULONG FormatVersion;
    ULONG TotalSize;
    ULONG DefaultTocOffset; // to ACTIVATION_CONTEXT_DATA_TOC_HEADER, from this struct base
    ULONG ExtendedTocOffset; // to ACTIVATION_CONTEXT_DATA_EXTENDED_TOC_HEADER, from this struct base; zero if absent
    ULONG AssemblyRosterOffset; // to ACTIVATION_CONTEXT_DATA_ASSEMBLY_ROSTER_HEADER, from this struct base
    ULONG Flags; // ACTIVATION_CONTEXT_FLAG_*
} ACTIVATION_CONTEXT_DATA, * PACTIVATION_CONTEXT_DATA;

typedef struct _ACTIVATION_CONTEXT
{
    LONG RefCount;
    ULONG Flags; // bit 0: zombified; bit 1: allocation is not freed by RtlpFreeActivationContext
    LIST_ENTRY Links;
    PACTIVATION_CONTEXT_DATA ActivationContextData;
    PVOID NotificationRoutine;
    PVOID NotificationContext;
    ULONG SentNotifications[8]; // notification-type bitmap
    ULONG DisabledNotifications[8]; // notification-type bitmap
    ASSEMBLY_STORAGE_MAP StorageMap;
    PASSEMBLY_STORAGE_MAP_ENTRY InlineStorageMapEntries[32];
    ULONG StackTraceIndex;
    PVOID StackTraces[4][4];
} ACTIVATION_CONTEXT, * PACTIVATION_CONTEXT;


typedef struct _RTL_ACTIVATION_CONTEXT_STACK_FRAME
{
    struct _RTL_ACTIVATION_CONTEXT_STACK_FRAME* Previous;
    PACTIVATION_CONTEXT ActivationContext;
    ULONG Flags; // RTL_ACTIVATION_CONTEXT_STACK_FRAME_FLAG_*
} RTL_ACTIVATION_CONTEXT_STACK_FRAME, * PRTL_ACTIVATION_CONTEXT_STACK_FRAME;

typedef struct _ACTIVATION_CONTEXT_STACK
{
    PRTL_ACTIVATION_CONTEXT_STACK_FRAME ActiveFrame;
    LIST_ENTRY FrameListCache;
    ULONG Flags; // ACTIVATION_CONTEXT_STACK_FLAG_*
    ULONG NextCookieSequenceNumber;
    ULONG StackId;
} ACTIVATION_CONTEXT_STACK, * PACTIVATION_CONTEXT_STACK;

#define GDI_BATCH_BUFFER_SIZE 310

typedef struct _GDI_TEB_BATCH
{
    ULONG Offset : 30;
    ULONG InProcessing : 1;
    ULONG HasRenderingCommand : 1;
    ULONG_PTR HDC;
    ULONG Buffer[GDI_BATCH_BUFFER_SIZE];
} GDI_TEB_BATCH, * PGDI_TEB_BATCH;


typedef struct _TEB
{
    NT_TIB NtTib;

    PVOID EnvironmentPointer;


    CLIENT_ID ClientId;
    PVOID ActiveRpcHandle;

    PVOID* ThreadLocalStoragePointer;

    PPEB ProcessEnvironmentBlock;

    ULONG LastErrorValue;


    ULONG CountOfOwnedCriticalSections;

    PVOID CsrClientThread;


    PVOID Win32ThreadInfo;


    ULONG User32Reserved[26];

    ULONG UserReserved[5];


    PVOID WOW32Reserved;
    LCID CurrentLocale;

    ULONG FpSoftwareStatusRegister;

    //
    // Reserved.
    //
    PVOID ReservedForDebuggerInstrumentation[16];

#ifdef _WIN64
    //
    // Reserved for floating-point emulation.
    //
    PVOID SystemReserved1[25];

    //
    // Per-thread fiber local storage. (Teb->HasFiberData)
    //
    PVOID HeapFlsData;

    //
    // Reserved.
    //
    ULONG_PTR RngState[4];
#else
    //
    // Reserved.
    //
    PVOID SystemReserved1[26];
#endif

    //
    // Placeholder compatibility mode. (ProjFs and Cloud Files)
    //
    CHAR PlaceholderCompatibilityMode;

    //
    // Indicates whether placeholder hydration is always explicit.
    //
    BOOLEAN PlaceholderHydrationAlwaysExplicit;

    //
    // ProjFs and Cloud Files (reparse point) file virtualization.
    //
    CHAR PlaceholderReserved[10];

    //
    // The process ID (PID) that the current COM server thread is acting on behalf of.
    //
    ULONG ProxiedProcessId;

    //
    // Pointer to the activation context stack for the current thread.
    //
    ACTIVATION_CONTEXT_STACK ActivationStack;

    //
    // Scheduler ticket attributing this thread's CPU usage to another thread. Set/cleared via NtSetInformationThread(ThreadWorkOnBehalfTicket).
    //
    UCHAR WorkingOnBehalfTicket[8]; // ALPC_WORK_ON_BEHALF_TICKET

    //
    // The last exception status for the current thread.
    //
    NTSTATUS ExceptionCode;

    //
    // Pointer to the activation context stack for the current thread.
    //
    PACTIVATION_CONTEXT_STACK ActivationContextStackPointer;

    //
    // The stack pointer (SP) of the current system call or exception during instrumentation.
    //
    ULONG_PTR InstrumentationCallbackSp;

    //
    // The program counter (PC) of the previous system call or exception during instrumentation.
    //
    ULONG_PTR InstrumentationCallbackPreviousPc;

    //
    // The stack pointer (SP) of the previous system call or exception during instrumentation.
    //
    ULONG_PTR InstrumentationCallbackPreviousSp;

#ifdef _WIN64
    //
    // The miniversion ID of the current transacted file operation.
    //
    ULONG TxFsContext;
#endif

    //
    // Indicates the state of the system call or exception instrumentation callback.
    //
    BOOLEAN InstrumentationCallbackDisabled;

#ifdef _WIN64
    //
    // Indicates the state of alignment exceptions for unaligned load/store operations.
    //
    BOOLEAN UnalignedLoadStoreExceptions;
#endif

#ifndef _WIN64
    //
    // SpareBytes.
    //
    UCHAR SpareBytes[23];

    //
    // The miniversion ID of the current transacted file operation.
    //
    ULONG TxFsContext;
#endif

    //
    // Reserved for GDI (Win32k).
    //
    GDI_TEB_BATCH GdiTebBatch;
    CLIENT_ID RealClientId;
    HANDLE GdiCachedProcessHandle;
    ULONG GdiClientPID;
    ULONG GdiClientTID;
    PVOID GdiThreadLocalInfo;



    //
    // Reserved for opengl32.dll
    //
    PVOID glDispatchTable[233];
    ULONG_PTR glReserved1[29];
    PVOID glReserved2;
    PVOID glSectionInfo;
    PVOID glSection;
    PVOID glTable;
    PVOID glCurrentRC;
    PVOID glContext;

    //
    // The previous status value for this thread.
    //
    NTSTATUS LastStatusValue;

    //
    // A static string for use by the application.
    //
    UNICODE_STRING StaticUnicodeString;

    //
    // A static buffer for use by the application.
    //
    WCHAR StaticUnicodeBuffer[261];

    //
    // The maximum stack size and indicates the base of the stack.
    //
    PVOID DeallocationStack;

    //
    // Data for Thread Local Storage. (TlsGetValue)
    //
    PVOID TlsSlots[TLS_MINIMUM_AVAILABLE];

    //
    // Reserved for TLS.
    //
    LIST_ENTRY TlsLinks;

    //
    // Reserved for NTVDM.
    //
    PVOID Vdm;

    //
    // Reserved for RPC. The pointer is XOR'ed with RPC_THREAD_POINTER_KEY.
    //
    PVOID ReservedForNtRpc;

    //
    // Reserved for Debugging (DebugActiveProcess).
    //
    PVOID DbgSsReserved[2];

    //
    // The error mode for the current thread. (GetThreadErrorMode)
    //
    ULONG HardErrorMode;

    //
    // Reserved.
    //
#ifdef _WIN64
    PVOID Instrumentation[11];
#else
    PVOID Instrumentation[9];
#endif

    //
    // Reserved.
    //
    GUID ActivityId;

    //
    // The identifier of the service that created the thread. (svchost)
    //
    PVOID SubProcessTag;

    //
    // Reserved.
    //
    PVOID PerflibData;

    //
    // Reserved.
    //
    PVOID EtwTraceData;

    //
    // The address of a socket handle during a blocking socket operation. (WSAStartup)
    //
    HANDLE WinSockData;

    //
    // The number of function calls accumulated in the current GDI batch. (GdiSetBatchLimit)
    //
    ULONG GdiBatchCount;

    //
    // The preferred processor for the current thread. (SetThreadIdealProcessor/SetThreadIdealProcessorEx)
    //
    union
    {
        PROCESSOR_NUMBER CurrentIdealProcessor;
        ULONG IdealProcessorValue;
        struct
        {
            UCHAR ReservedPad0;
            UCHAR ReservedPad1;
            UCHAR ReservedPad2;
            UCHAR IdealProcessor;
        };
    };

    //
    // The minimum size of the stack available during any stack overflow exceptions. (SetThreadStackGuarantee)
    //
    ULONG GuaranteedStackBytes;

    //
    // Reserved.
    //
    PVOID ReservedForPerf;

    //
    // Reserved for Object Linking and Embedding (OLE)
    //
    //PSOleTlsData ReservedForOle;

    //
    // Indicates whether the thread is waiting on the loader lock.
    //
    ULONG WaitingOnLoaderLock;

    //
    // The saved priority state for the thread.
    //
    PVOID SavedPriorityState;

    //
    // Reserved.
    //
    ULONG_PTR ReservedForCodeCoverage;

    //
    // Per-thread state for the NT thread pool.
    //
   // PTPP_THREAD_DATA ThreadPoolData;

    //
    // Pointer to the TLS (Thread Local Storage) expansion slots for the thread.
    //
    PVOID* TlsExpansionSlots;

#ifdef _WIN64
    PVOID ChpeV2CpuAreaInfo; // CHPEV2_CPUAREA_INFO // previously DeallocationBStore
    PVOID Unused; // previously BStoreLimit
#endif

    //
    // The generation of the MUI (Multilingual User Interface) data.
    //
    ULONG MuiGeneration;

    //
    // Indicates whether the thread is impersonating another security context.
    //
    ULONG IsImpersonating;

    //
    // Pointer to the NLS (National Language Support) cache.
    //
    PVOID NlsCache;

    //
    // Pointer to the AppCompat/Shim Engine data.
    //
    PVOID pShimData;

    //
    // Reserved.
    //
    ULONG HeapData;

    //
    // Handle to the current transaction associated with the thread.
    //
    HANDLE CurrentTransactionHandle;

    //
    // Pointer to the active frame for the thread.
    //
    //PTEB_ACTIVE_FRAME ActiveFrame;

    //
    // Reserved for FLS (RtlProcessFlsData).
    //
    PVOID FlsData;

    //
    // Pointer to the preferred languages for the current thread. (GetThreadPreferredUILanguages)
    //
    PVOID PreferredLanguages;

    //
    // Pointer to the user-preferred languages for the current thread. (GetUserPreferredUILanguages)
    //
    PVOID UserPrefLanguages;

    //
    // Pointer to the merged preferred languages for the current thread. (MUI_MERGE_USER_FALLBACK)
    //
    PVOID MergedPrefLanguages;

    //
    // Indicates whether the thread is impersonating another user's language settings.
    //
    ULONG MuiImpersonation;

    //
    // Reserved.
    //
    union
    {
        USHORT CrossTebFlags;
        USHORT SpareCrossTebBits : 16;
    };

    //
    // SameTebFlags modify the state and behavior of the current thread.
    //
    union
    {
        USHORT SameTebFlags;
        struct
        {
            USHORT SafeThunkCall : 1;
            USHORT InDebugPrint : 1;            // Indicates if the thread is currently in a debug print routine.
            USHORT HasFiberData : 1;            // Indicates if the thread has local fiber-local storage (FLS).
            USHORT SkipThreadAttach : 1;        // Indicates if the thread should suppress DLL_THREAD_ATTACH notifications.
            USHORT WerInShipAssertCode : 1;
            USHORT RanProcessInit : 1;          // Indicates if the thread has run process initialization code.
            USHORT ClonedThread : 1;            // Indicates if the thread is a clone of a different thread.
            USHORT SuppressDebugMsg : 1;        // Indicates if the thread should suppress LOAD_DLL_DEBUG_INFO notifications.
            USHORT DisableUserStackWalk : 1;
            USHORT RtlExceptionAttached : 1;
            USHORT InitialThread : 1;           // Indicates if the thread is the initial thread of the process.
            USHORT SessionAware : 1;
            USHORT LoadOwner : 1;               // Indicates if the thread is the owner of the process loader lock.
            USHORT LoaderWorker : 1;
            USHORT SkipLoaderInit : 1;
            USHORT SkipFileAPIBrokering : 1;
        };
    };

    //
    // Pointer to the callback function that is called when a KTM transaction scope is entered.
    //
    PVOID TxnScopeEnterCallback;

    //
    // Pointer to the callback function that is called when a KTM transaction scope is exited.
    ///
    PVOID TxnScopeExitCallback;

    //
    // Pointer to optional context data for use by the application when a KTM transaction scope callback is called.
    //
    PVOID TxnScopeContext;

    //
    // The lock count of critical sections for the current thread.
    //
    ULONG LockCount;

    //
    // The offset to the WOW64 (Windows on Windows) TEB for the current thread.
    //
    LONG WowTebOffset;

    //
    // Pointer to the DLL containing the resource (valid after LdrFindResource_U/LdrResFindResource/etc... returns).
    //
    PVOID ResourceRetValue;

    //
    // Reserved for Windows Driver Framework (WDF).
    //
    PVOID ReservedForWdf;

    //
    // Reserved for the Microsoft C runtime (CRT).
    //
    ULONGLONG ReservedForCrt;

    //
    // The Host Compute Service (HCS) container identifier.
    //
    GUID EffectiveContainerId;

    //
    // Reserved for Kernel32!Sleep (SpinWait).
    //
    ULONGLONG LastSleepCounter; // since Win11

    //
    // Reserved for Kernel32!Sleep (SpinWait).
    //
    ULONG SpinCallCount;

    //
    // Extended feature disable mask (AVX).
    //
    union
    {
        ULONGLONG ExtendedFeatureDisableMask;
        struct
        {
            ULONGLONG LegacyFloatingPointDisabled : 1;
            ULONGLONG LegacySseDisabled : 1;
            ULONGLONG AvxDisabled : 1;
            ULONGLONG MpxBndregsDisabled : 1;
            ULONGLONG MpxBndcsrDisabled : 1;
            ULONGLONG Avx512OpmaskDisabled : 1;
            ULONGLONG Avx512ZmmHi256Disabled : 1;
            ULONGLONG Avx512Zmm16HiDisabled : 1;
            ULONGLONG IptDisabled : 1;
            ULONGLONG PkruDisabled : 1;
            ULONGLONG Reserved0 : 1;
            ULONGLONG CetUserDisabled : 1;
            ULONGLONG CetSupervisorDisabled : 1;
            ULONGLONG Reserved1 : 4;
            ULONGLONG AmxTileDisabled : 1;
            ULONGLONG AmxBf16Disabled : 1;
            ULONGLONG AmxInt8Disabled : 1;
            ULONGLONG Reserved2 : 44;
        };
    };

    //
    // Reserved.
    //
    PVOID SchedulerSharedDataSlot; // since 24H2

    //
    // Reserved.
    //
    PVOID HeapWalkContext;

    //
    // The primary processor group affinity of the thread.
    //
    GROUP_AFFINITY PrimaryGroupAffinity;

    //
    // Read-copy-update (RCU) synchronization context.
    //
    ULONG Rcu[2];
} TEB, * PTEB;