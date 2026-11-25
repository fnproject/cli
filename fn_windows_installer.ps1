function Set-EnvironmentVariable
{
  param
  (
    [Parameter(Mandatory=$true)]
    [String]
    $value,
    
    [Parameter(Mandatory=$true)]
    [EnvironmentVariableTarget]
    $target
  )
    WriteLog "INFO" "-- Arguments passed --"
    WriteLog "INFO" "value[$value]"
    WriteLog "INFO" "target[$target]"
    
    $path = [Environment]::GetEnvironmentVariable('Path', $target)
	
	WriteLog "INFO" "Old Path[$path]"
	
	if( $PATH -notlike "*;"+$value+";*" ){
        $newpath = $path + ';'+$value+';'
		[Environment]::SetEnvironmentVariable("Path", $newpath, $target)
    }
	$NewPath = ([Environment]::GetEnvironmentVariable('Path', $target))
    WriteLog "INFO" "New Path[$NewPath]"
}

Function WriteLog {
    [CmdletBinding()]
    Param(
    [Parameter(Mandatory=$False)]
    [ValidateSet("INFO","WARN","ERROR","VERBOSE","DEBUG")]
    [String]
    $Level = "INFO",

    [Parameter(Mandatory=$True)]
    [string]
    $Message
	
    )

    $Stamp = (Get-Date).toString("yyyy/MM/dd HH:mm:ss")
    $Line = "$Stamp $Level $Message"
	
	Add-Content $Logfile -Value $Line

    if($Level -eq "INFO") {
        Write-Information  -Message $Line -InformationAction Continue
    } elseif($Level -eq "WARN"){
		 Write-Warning  -Message $Line
	} elseif($Level -eq "ERROR"){
		 Write-Error  -Message $Line
	} elseif($Level -eq "VERBOSE"){
		Write-Verbose -Message $Line  -Verbose
	} elseif($Level -eq "DEBUG"){
		 Write-Debug  -Message $Line -Debug
	}else{
		 Write-Information  -Message $Line -InformationAction Continue
	}
}

function DownloadFile
{
  param
  (
    [Parameter(Mandatory=$true)]
    [String]
    $Path,
	 [Parameter(Mandatory=$true)]
    [String]
    $FileName,
    [Parameter(Mandatory=$true)]
    [String]
    $URL
   
  )
    $MaxRetries = 3 
    $RetryCount = 0
    $SleepTime = 2

    $null = mkdir $Path -Force
    
    $StartBitsTransferSplat = @{
        Source        = $URL 
        Destination   = ($Path + $FileName) 
        RetryInterval = 60
    }
    while ($RetryCount -le $MaxRetries){
        Try {
			$Loc = ($Path+$FileName)
            WriteLog "VERBOSE" "Trying to download file from URL[$URL] at location[$Loc]"
            
            $null = if((Get-Service BITS).Status -eq "Running") {
                Start-BitsTransfer @StartBitsTransferSplat -ErrorAction Stop
            } else {
                Invoke-WebRequest $URL -OutFile ($Path + $FileName)
            }
           
            # To get here, the transfer must have finished, so set the counter
            # greater than the max value to exit the loop
            $RetryCount = $MaxRetries + 1
        } # End Try block
        Catch {
            WriteLog "WARN" "Download of file from URL[$URL] failed."
			WriteLog "WARN" "Going to sleep for [$SleepTime] second and then will try again to download file[$URL]"
			
            Start-Sleep -Seconds $SleepTime 

            $PSItem.Exception.Message
            $retryCount += 1
            $SleepTime = 2 * $SleepTime
            if($RetryCount -gt $MaxRetries){
				WriteLog "ERROR" "Reached maximum no of retries while downloading file from URL[$URL]"
				Throw
			}
            WriteLog "WARN" "Attempting retry #: $RetryCount"
        } # End Catch Block
    } # End While loop for retries
	WriteLog "VERBOSE" "Downloaded file from URL[$URL] at location[($Path$FileName)]"
	
}

function ShowSytemsConfigState
{	
	try{
		
		$OsName=(Get-WmiObject Win32_OperatingSystem).caption
        $OSArchitecture=((Get-WmiObject Win32_OperatingSystem).OSArchitecture).ToLower()
        $OsVersion=(Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").ReleaseId

        $OsBuildDetails = Get-ComputerInfo OsName,OsVersion,OsBuildNumber,OsHardwareAbstractionLayer,WindowsVersion

        $RamInGB =(Get-CimInstance Win32_PhysicalMemory | Measure-Object -Property capacity -Sum).sum /1gb

        $HyperVRequirementVirtualizationFirmwareEnabled=Get-ComputerInfo -property "HyperVRequirementVirtualizationFirmwareEnabled" | Select -expand HyperVRequirementVirtualizationFirmwareEnabled

        $HyperVisorPresent=Get-ComputerInfo -property "HyperVisorPresent" | Select -expand HyperVisorPresent

        $SupportedOsNamesStr = $SupportedOsNames -join ','

        WriteLog "INFO" "#########################################"
        WriteLog "INFO" "#########################################"
        
        WriteLog "INFO" "OS Name[$OsName]"

        WriteLog "INFO" "OsBuildDetails[$OsBuildDetails]"
        
        WriteLog "INFO" "Current System OS Architecture[$OSArchitecture]"

        WriteLog "INFO" "Current System RAM size[$RamInGB GB] "
                
        
        if($HyperVRequirementVirtualizationFirmwareEnabled -eq $null -or $HyperVRequirementVirtualizationFirmwareEnabled){
            WriteLog "INFO" "Current system supports virtualization at Firmware(BIOS) level"
        }else{
            WriteLog "ERROR" "Current system does not support virtualization at Firmware(BIOS) level. Please enable it at BIOS level."
        }
                
		if ((Get-CimInstance Win32_OperatingSystem).Caption -match 'Microsoft Windows 10')
		{
			#WriteLog "INFO" "Machine os[Microsoft Windows 10]";
			if ((Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online).State -ne 'Enabled')
			{
				WriteLog "INFO" "Microsoft Windows 10 Hyper-V is not enabled"
				
			}else{
				WriteLog "INFO" "Microsoft Windows 10 Hyper-V is already enabled";
			}
		}elseif ((Get-CimInstance Win32_OperatingSystem).Caption -match 'Microsoft Windows Server')
		{
			#WriteLog "INFO" "Machine os[Microsoft Windows Server]";
			if ((Get-WindowsFeature -Name Hyper-V) -eq $false)
			{
				WriteLog "INFO" "Microsoft Windows Server Hyper-V is not enabled";
				
			}else{
				WriteLog "INFO" "Microsoft Windows Server Hyper-V is already enabled";
			}
		}

	} Catch {
		Throw
	}
}

function EnableHyperviser
{
	try{
		WriteLog "INFO" "Verifying Hyper-V in machine..."

		$oldPolicy =Get-ExecutionPolicy

		WriteLog "INFO" "Currrent Execution Policy[$oldPolicy]"
		
		$Msg = ((Get-CimInstance Win32_OperatingSystem).Caption)
		WriteLog "INFO" "Windows 10 Details[$Msg]"

		if ((Get-CimInstance Win32_OperatingSystem).Caption -match 'Microsoft Windows 10')
		{
			WriteLog "INFO" "Machine os[Microsoft Windows 10]";
			if ((Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online).State -ne 'Enabled')
			{
				WriteLog "INFO" "Started enabling Microsoft Windows 10 Hyper-V-All"
				
				Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
				
				WriteLog "INFO" "Enabled Microsoft Windows 10 Hyper-V-All";
			}else{
				WriteLog "INFO" "Microsoft Windows 10 Hyper-V-All is already enabled";
			}
		}elseif ((Get-CimInstance Win32_OperatingSystem).Caption -match 'Microsoft Windows Server')
		{
			WriteLog "INFO" "Machine os[Microsoft Windows Server]";
			if ((Get-WindowsFeature -Name Hyper-V) -eq $false)
			{
				WriteLog "INFO" "Started enabling Microsoft Windows Server Hyper-V";
				Install-WindowsFeature -Name Hyper-V -IncludeManagementTools
				WriteLog "INFO" "Enabled Microsoft Windows Server Hyper-V";
			}else{
				WriteLog "INFO" "Microsoft Windows Server Hyper-V is already enabled";
			}
		}

	} Catch {
		Throw
	}
}

function InstallFnClient
{
	try
	{
		
		WriteLog "INFO" "Determining latest release"
		
		$releases = "https://api.github.com/repos/fnproject/cli/releases"
		
		$tag = (Invoke-WebRequest -Uri $releases -UseBasicParsing | ConvertFrom-Json)[0].tag_name

		$fn_client_download_url = "https://github.com/fnproject/cli/releases/download/$tag/fn.exe"

		WriteLog "INFO" "Starting fn client download from url[$fn_client_download_url]"

		$fn_client_install_dir=$InstallPath+"\fn\"
		
		$fn_client_name='fn.exe'
				
		$null = mkdir $fn_client_install_dir -Force
		
		WriteLog "INFO" "fn client install dir[$fn_client_install_dir]"
		
		WriteLog "INFO" "fn client install Log File location[$Logfile]"

		#Downloading
		DownloadFile "$fn_client_install_dir" "$fn_client_name" "$fn_client_download_url" 

		WriteLog "INFO" "Downloaded fn client at location[$fn_client_install_dir]"

		WriteLog "INFO" "Started Setting path to include fn client executable"
		
		# target can be User or Machine
		#$target='User'
		Set-EnvironmentVariable -value $fn_client_install_dir -target User
		
		#WriteLog "INFO" "New Path[$NewPath]"
	
		WriteLog "INFO" "Ended Setting path to include fn client executable"
		

	} Catch {
		Throw
	}
}

function Install
{	
	WriteLog "INFO" "Performing Task[$Task] execution"
	WriteLog "INFO" "Installation Path[$InstallPath]"
	WriteLog "INFO" "Log file[$Logfile]"
	    
	try{
        if($Task -eq "get-system-state") {
			ShowSytemsConfigState
  		}elseif($Task -eq "fn-client-install"){
			InstallFnClient
		} elseif($Task -eq "enable-hyperviser"){
			EnableHyperviser
		}else{
            WriteLog "ERROR" "Invalid argument provided[$Task]."
            Throw
		}
	} Catch {
		Throw
	}
}


######################################################################################################
#####################Main Start#######################################################################
######################################################################################################

$Task=$args[0]

#Configuring all directories required for docker install
$InstallPath = "c:\fn_install"
$LogPath = $InstallPath+"\logs"
$null = mkdir $LogPath -Force
$Logfile = ($LogPath + "\fn_install.log")

#Trigering installation
Install