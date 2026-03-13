@echo off
setlocal enabledelayedexpansion
set FOUND=0
for /f "tokens=2 delims==" %%p in ('wmic process where "CommandLine like '%%gprobe%%networkid 8004%%'" get ProcessId /format:list 2^>nul ^| findstr ProcessId') do (
  taskkill /PID %%p /F >nul 2>&1
  echo Node stopped ^(PID: %%p^)
  set FOUND=1
)
if !FOUND!==0 (
  echo No running node found.
)
endlocal
