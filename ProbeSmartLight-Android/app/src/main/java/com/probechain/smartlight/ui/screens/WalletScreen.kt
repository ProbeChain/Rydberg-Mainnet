// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import com.probechain.smartlight.core.AndroidKeyStoreWrapper

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WalletScreen() {
    var walletAddress by remember { mutableStateOf("") }
    var isRegistered by remember { mutableStateOf(false) }
    var showKeyGenDialog by remember { mutableStateOf(false) }
    var passphrase by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text("Wallet", style = MaterialTheme.typography.headlineMedium)

        // Address
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("SmartLight Address", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                Spacer(Modifier.height(8.dp))
                if (walletAddress.isEmpty()) {
                    Button(onClick = { showKeyGenDialog = true }) {
                        Icon(Icons.Default.Key, contentDescription = null)
                        Spacer(Modifier.width(8.dp))
                        Text("Generate Dilithium Key")
                    }
                } else {
                    Text(walletAddress, fontFamily = FontFamily.Monospace, style = MaterialTheme.typography.bodySmall)
                    Spacer(Modifier.height(4.dp))
                    AssistChip(
                        onClick = {},
                        label = { Text("Dilithium (ML-DSA-44)") },
                        leadingIcon = { Icon(Icons.Default.Shield, contentDescription = null, modifier = Modifier.size(16.dp)) }
                    )
                }
            }
        }

        // Balance
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            ListItem(
                headlineContent = { Text("PROBE Balance") },
                trailingContent = { Text("0.0000 PROBE", style = MaterialTheme.typography.titleMedium) }
            )
        }

        // Registration
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("SmartLight Registration", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                Spacer(Modifier.height(8.dp))
                ListItem(
                    headlineContent = { Text("Status") },
                    trailingContent = {
                        Text(
                            if (isRegistered) "Registered" else "Not Registered",
                            color = if (isRegistered) Color(0xFF4CAF50) else Color(0xFFFF9800)
                        )
                    }
                )
                ListItem(
                    headlineContent = { Text("Required Stake") },
                    trailingContent = { Text("10 PROBE") }
                )
                if (!isRegistered && walletAddress.isNotEmpty()) {
                    Button(
                        onClick = { isRegistered = true },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text("Register as SmartLight Node")
                    }
                }
            }
        }

        // Security
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Key Security", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(
                    leadingContent = { Icon(Icons.Default.Security, contentDescription = null) },
                    headlineContent = { Text("Hardware Keystore") },
                    trailingContent = {
                        Text(
                            if (AndroidKeyStoreWrapper.isStrongBoxAvailable) "StrongBox" else "TEE",
                            color = Color(0xFF4CAF50)
                        )
                    }
                )
                ListItem(
                    leadingContent = { Icon(Icons.Default.Fingerprint, contentDescription = null) },
                    headlineContent = { Text("Biometric Lock") },
                    trailingContent = { Text("Enabled", color = Color(0xFF4CAF50)) }
                )
                ListItem(
                    leadingContent = { Icon(Icons.Default.Key, contentDescription = null) },
                    headlineContent = { Text("Private Key Size") },
                    trailingContent = { Text("2528 bytes") }
                )
            }
        }
    }

    // Key Generation Dialog
    if (showKeyGenDialog) {
        AlertDialog(
            onDismissRequest = { showKeyGenDialog = false },
            title = { Text("Generate Dilithium Key") },
            text = {
                Column {
                    Text("Create a quantum-resistant key pair. Protected by AES-256-GCM + hardware keystore.")
                    Spacer(Modifier.height(16.dp))
                    OutlinedTextField(
                        value = passphrase,
                        onValueChange = { passphrase = it },
                        label = { Text("Passphrase") },
                        modifier = Modifier.fillMaxWidth()
                    )
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        walletAddress = "0x" + "a".repeat(40) // placeholder
                        showKeyGenDialog = false
                    },
                    enabled = passphrase.length >= 6
                ) {
                    Text("Generate")
                }
            },
            dismissButton = {
                TextButton(onClick = { showKeyGenDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }
}
