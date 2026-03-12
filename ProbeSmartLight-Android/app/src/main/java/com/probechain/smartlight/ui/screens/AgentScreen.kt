// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.probechain.smartlight.service.NodeViewModel

@Composable
fun AgentScreen(viewModel: NodeViewModel = viewModel()) {
    val isRunning by viewModel.isRunning.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text("Agent", style = MaterialTheme.typography.headlineMedium)

        // Task Runner Status
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Agent Task Runner", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(
                    headlineContent = { Text("Status") },
                    trailingContent = {
                        Text(
                            if (isRunning) "Active" else "Inactive",
                            color = if (isRunning) Color(0xFF4CAF50) else Color.Gray
                        )
                    }
                )
                ListItem(
                    headlineContent = { Text("Max Concurrent Tasks") },
                    trailingContent = { Text("2") }
                )
                ListItem(
                    headlineContent = { Text("Max Memory per Task") },
                    trailingContent = { Text("25 MB") }
                )
                ListItem(
                    headlineContent = { Text("Task Timeout") },
                    trailingContent = { Text("5 seconds") }
                )
            }
        }

        // Statistics
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Statistics", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(
                    headlineContent = { Text("Active Tasks") },
                    trailingContent = { Text("0 / 2", color = Color(0xFF2196F3)) }
                )
                ListItem(
                    headlineContent = { Text("Tasks Completed") },
                    trailingContent = { Text("0") }
                )
                ListItem(
                    headlineContent = { Text("Success Rate") },
                    trailingContent = { Text("100.0%", color = Color(0xFF4CAF50)) }
                )
            }
        }

        // Power Mode Impact
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Power Mode Impact", style = MaterialTheme.typography.titleSmall, color = Color.Gray)
                ListItem(
                    leadingContent = { Icon(Icons.Default.BoltCircle, contentDescription = null, tint = Color(0xFFFFC107)) },
                    headlineContent = { Text("Full Mode") },
                    trailingContent = { Text("Tasks Enabled", color = Color(0xFF4CAF50), style = MaterialTheme.typography.bodySmall) }
                )
                ListItem(
                    leadingContent = { Icon(Icons.Default.Eco, contentDescription = null, tint = Color(0xFF4CAF50)) },
                    headlineContent = { Text("Eco Mode") },
                    trailingContent = { Text("Tasks Disabled", color = Color(0xFFFF9800), style = MaterialTheme.typography.bodySmall) }
                )
                ListItem(
                    leadingContent = { Icon(Icons.Default.Bedtime, contentDescription = null, tint = Color(0xFF3F51B5)) },
                    headlineContent = { Text("Sleep Mode") },
                    trailingContent = { Text("Tasks Disabled", color = Color(0xFFF44336), style = MaterialTheme.typography.bodySmall) }
                )
            }
        }

        Text(
            "Agent tasks are lightweight computations requested by the ProbeChain network. " +
            "Tasks execute in a sandboxed environment with strict resource limits. " +
            "Only available in Full power mode (while charging).",
            style = MaterialTheme.typography.bodySmall,
            color = Color.Gray
        )
    }
}
