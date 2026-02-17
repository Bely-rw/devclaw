@echo off
:: DevClaw installer — Windows CMD wrapper
:: Launches the PowerShell installer
powershell -ExecutionPolicy Bypass -Command "& { iwr -useb https://raw.githubusercontent.com/jholhewres/devclaw/main/scripts/install/install.ps1 | iex }"
