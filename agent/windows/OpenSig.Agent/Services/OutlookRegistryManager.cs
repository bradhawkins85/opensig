using System;
using Microsoft.Win32;

namespace OpenSig.Agent.Services
{
    public class OutlookRegistryManager
    {
        private const string OutlookProfilesKeyPath = @"Software\Microsoft\Office\16.0\Outlook\Profiles\Outlook";
        
        public OutlookRegistryManager()
        {
        }

        /// <summary>
        /// Sets the default signature for new emails and replies/forwards in Outlook
        /// </summary>
        /// <param name="signatureName">Name of the signature to set as default (without extension)</param>
        /// <returns>True if successful, false otherwise</returns>
        public bool SetDefaultSignatures(string signatureName)
        {
            try
            {
                Logger.Log($"Setting default Outlook signatures to: {signatureName}");

                // Try multiple Outlook versions (16.0 = Office 2016/2019/365, 15.0 = Office 2013, 14.0 = Office 2010)
                string[] outlookVersions = { "16.0", "15.0", "14.0" };
                bool anySuccess = false;

                foreach (var version in outlookVersions)
                {
                    string profilesPath = $@"Software\Microsoft\Office\{version}\Outlook\Profiles\Outlook";
                    
                    if (TrySetDefaultSignaturesForVersion(profilesPath, signatureName))
                    {
                        Logger.Log($"Successfully set default signatures for Office {version}");
                        anySuccess = true;
                    }
                }

                if (!anySuccess)
                {
                    Logger.LogWarning("Could not set default signatures for any Outlook version. This may be because:");
                    Logger.LogWarning("  1. Outlook is not installed or hasn't been run yet");
                    Logger.LogWarning("  2. Roaming signatures are enabled (managed by Exchange/Microsoft 365)");
                    Logger.LogWarning("  3. The user profile hasn't been created");
                }

                return anySuccess;
            }
            catch (Exception ex)
            {
                Logger.LogError($"Error setting default Outlook signatures: {ex.Message}");
                return false;
            }
        }

        private bool TrySetDefaultSignaturesForVersion(string profilesPath, string signatureName)
        {
            try
            {
                using (RegistryKey? profilesKey = Registry.CurrentUser.OpenSubKey(profilesPath, writable: false))
                {
                    if (profilesKey == null)
                    {
                        return false;
                    }

                    // Look for account keys (they contain account configuration)
                    string[] subKeyNames = profilesKey.GetSubKeyNames();
                    bool foundAnyAccount = false;

                    foreach (string subKeyName in subKeyNames)
                    {
                        // Account keys are typically long hex strings
                        if (subKeyName.Length >= 16)
                        {
                            string accountKeyPath = $@"{profilesPath}\{subKeyName}";
                            
                            if (SetSignatureForAccount(accountKeyPath, signatureName))
                            {
                                foundAnyAccount = true;
                            }
                        }
                    }

                    return foundAnyAccount;
                }
            }
            catch (Exception ex)
            {
                Logger.LogWarning($"Could not access registry path {profilesPath}: {ex.Message}");
                return false;
            }
        }

        private bool SetSignatureForAccount(string accountKeyPath, string signatureName)
        {
            try
            {
                using (RegistryKey? accountKey = Registry.CurrentUser.OpenSubKey(accountKeyPath, writable: true))
                {
                    if (accountKey == null)
                    {
                        return false;
                    }

                    // Check if roaming signatures are enabled (this registry value would be present if managed)
                    object? roamingValue = accountKey.GetValue("RoamingSigs");
                    if (roamingValue != null && Convert.ToInt32(roamingValue) == 1)
                    {
                        Logger.LogWarning($"Roaming signatures are enabled for this account. Skipping default signature configuration.");
                        return false;
                    }

                    // Set default signature for new emails
                    // Value type: REG_BINARY (byte array of Unicode string)
                    byte[] signatureBytes = System.Text.Encoding.Unicode.GetBytes(signatureName + "\0");
                    accountKey.SetValue("New Signature", signatureBytes, RegistryValueKind.Binary);
                    
                    // Set default signature for replies/forwards
                    accountKey.SetValue("Reply-Forward Signature", signatureBytes, RegistryValueKind.Binary);
                    
                    Logger.Log($"Set default signatures for account at {accountKeyPath}");
                    return true;
                }
            }
            catch (Exception ex)
            {
                Logger.LogWarning($"Could not set signature for account {accountKeyPath}: {ex.Message}");
                return false;
            }
        }

        /// <summary>
        /// Checks if roaming signatures are enabled for any Outlook profile
        /// </summary>
        /// <returns>True if roaming signatures are detected</returns>
        public bool IsRoamingSignaturesEnabled()
        {
            try
            {
                string[] outlookVersions = { "16.0", "15.0", "14.0" };

                foreach (var version in outlookVersions)
                {
                    string profilesPath = $@"Software\Microsoft\Office\{version}\Outlook\Profiles\Outlook";
                    
                    using (RegistryKey? profilesKey = Registry.CurrentUser.OpenSubKey(profilesPath, writable: false))
                    {
                        if (profilesKey == null)
                        {
                            continue;
                        }

                        string[] subKeyNames = profilesKey.GetSubKeyNames();
                        foreach (string subKeyName in subKeyNames)
                        {
                            if (subKeyName.Length >= 16)
                            {
                                string accountKeyPath = $@"{profilesPath}\{subKeyName}";
                                using (RegistryKey? accountKey = Registry.CurrentUser.OpenSubKey(accountKeyPath, writable: false))
                                {
                                    if (accountKey != null)
                                    {
                                        object? roamingValue = accountKey.GetValue("RoamingSigs");
                                        if (roamingValue != null && Convert.ToInt32(roamingValue) == 1)
                                        {
                                            return true;
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                return false;
            }
            catch (Exception ex)
            {
                Logger.LogWarning($"Error checking for roaming signatures: {ex.Message}");
                return false;
            }
        }
    }
}
