// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.probechain.smartlight.core.AndroidKeyStoreWrapper
import com.probechain.smartlight.service.NodeViewModel

@Composable
fun SettingsScreen(viewModel: NodeViewModel = viewModel()) {
    var gnssEnabled by remember { mutableStateOf(true) }
    var selectedPowerMode by remember { mutableIntStateOf(0) }
    var maxAgentTasks by remember { mutableIntStateOf(2) }
    var showResetDialog by remember { mutableStateOf(false) }

    val powerModes = listOf("Full (Charging)", "Eco (Battery > 30%)", "Sleep (Battery < 15%)")

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text("Settings", style = MaterialTheme.typography.headlineMedium)

        // Node
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Node", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(headlineContent = { Text("Network ID") }, trailingContent = { Text("1205 (ProbeChain)") })
                ListItem(headlineContent = { Text("Database Cache") }, trailingContent = { Text("64 MB") })
                ListItem(headlineContent = { Text("Max Peers") }, trailingContent = { Text("50") })
            }
        }

        // Power Management
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Power Management", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                Spacer(Modifier.height(8.dp))
                powerModes.forEachIndexed { index, label ->
                    Row(modifier = Modifier.fillMaxWidth()) {
                        RadioButton(
                            selected = selectedPowerMode == index,
                            onClick = {
                                selectedPowerMode = index
                                viewModel.setPowerMode(index)
                            }
                        )
                        Text(label, modifier = Modifier.padding(top = 12.dp))
                    }
                }
                Spacer(Modifier.height(8.dp))
                val usage = when (selectedPowerMode) {
                    0 -> "~5% / hour" to Color(0xFFFF9800)
                    1 -> "~2% / hour" to Color(0xFF4CAF50)
                    else -> "~0.5% / hour" to Color(0xFF4CAF50)
                }
                Text("Estimated Battery: ${usage.first}", color = usage.second, style = MaterialTheme.typography.bodySmall)
            }
        }

        // GNSS
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("GNSS Time Source", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    Text("Enable GPS Time Sampling", modifier = Modifier.padding(top = 8.dp))
                    Switch(checked = gnssEnabled, onCheckedChange = { gnssEnabled = it })
                }
                if (gnssEnabled) {
                    ListItem(headlineContent = { Text("Sample Interval") }, trailingContent = { Text("30 seconds") })
                    ListItem(headlineContent = { Text("Accuracy") }, trailingContent = { Text("~100 ns") })
                }
            }
        }

        // Agent
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Agent Tasks", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(headlineContent = { Text("Max Tasks") }, trailingContent = { Text("$maxAgentTasks") })
                ListItem(headlineContent = { Text("Max Memory per Task") }, trailingContent = { Text("25 MB") })
                ListItem(headlineContent = { Text("Task Timeout") }, trailingContent = { Text("5 seconds") })
            }
        }

        // Security
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Security", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(
                    headlineContent = { Text("Quantum-Safe Keys") },
                    trailingContent = { Text("CRYSTALS-Dilithium", color = Color(0xFF2196F3), style = MaterialTheme.typography.bodySmall) }
                )
                ListItem(
                    headlineContent = { Text("Key Protection") },
                    trailingContent = {
                        Text(
                            if (AndroidKeyStoreWrapper.isStrongBoxAvailable) "StrongBox + AES-256-GCM" else "TEE + AES-256-GCM",
                            color = Color(0xFF2196F3),
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                )
                ListItem(headlineContent = { Text("Anti-Sybil Stake") }, trailingContent = { Text("10 PROBE") })
            }
        }

        // About
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("About", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(headlineContent = { Text("Version") }, trailingContent = { Text("2.0.0-smartlight") })
                ListItem(headlineContent = { Text("ProbeChain") }, trailingContent = { Text("V2.0 PoB + StellarSpeed") })
            }
        }

        // Reset
        OutlinedButton(
            onClick = { showResetDialog = true },
            modifier = Modifier.fillMaxWidth(),
            colors = ButtonDefaults.outlinedButtonColors(contentColor = Color(0xFFF44336))
        ) {
            Text("Reset Node Data")
        }
    }

    if (showResetDialog) {
        AlertDialog(
            onDismissRequest = { showResetDialog = false },
            title = { Text("Reset Node Data") },
            text = { Text("This will delete all synced data and require re-sync. Your keys will be preserved.") },
            confirmButton = {
                TextButton(onClick = { showResetDialog = false }) {
                    Text("Reset", color = Color(0xFFF44336))
                }
            },
            dismissButton = {
                TextButton(onClick = { showResetDialog = false }) { Text("Cancel") }
            }
        )
    }
}
