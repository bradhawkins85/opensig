using System;
using System.Diagnostics;

namespace OpenSig.Agent.Services
{
    public static class Logger
    {
        private const string EventSourceName = "OpenSig.Agent";
        private const string EventLogName = "Application";
        private static bool _eventLogAvailable = false;

        static Logger()
        {
            // Check if we can write to Windows Event Log
            try
            {
                if (!EventLog.SourceExists(EventSourceName))
                {
                    // Creating event source requires admin privileges
                    // In production, this should be done during installation
                    _eventLogAvailable = false;
                }
                else
                {
                    _eventLogAvailable = true;
                }
            }
            catch (Exception)
            {
                _eventLogAvailable = false;
            }
        }

        public static void Log(string message)
        {
            // Always log to console
            Console.WriteLine($"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] INFO: {message}");

            // Also log to Windows Event Log if available
            if (_eventLogAvailable)
            {
                try
                {
                    EventLog.WriteEntry(EventSourceName, message, EventLogEntryType.Information);
                }
                catch (Exception)
                {
                    // Silently fail if we can't write to event log
                }
            }
        }

        public static void LogError(string message)
        {
            // Always log to console
            Console.Error.WriteLine($"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] ERROR: {message}");

            // Also log to Windows Event Log if available
            if (_eventLogAvailable)
            {
                try
                {
                    EventLog.WriteEntry(EventSourceName, message, EventLogEntryType.Error);
                }
                catch (Exception)
                {
                    // Silently fail if we can't write to event log
                }
            }
        }

        public static void LogWarning(string message)
        {
            // Always log to console
            Console.WriteLine($"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] WARNING: {message}");

            // Also log to Windows Event Log if available
            if (_eventLogAvailable)
            {
                try
                {
                    EventLog.WriteEntry(EventSourceName, message, EventLogEntryType.Warning);
                }
                catch (Exception)
                {
                    // Silently fail if we can't write to event log
                }
            }
        }
    }
}
