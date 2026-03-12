// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui.screens

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.probechain.smartlight.service.NodeViewModel

@Composable
fun DashboardScreen(viewModel: NodeViewModel = viewModel()) {
    val isRunning by viewModel.isRunning.collectAsState()
    val syncedBlock by viewModel.syncedBlock.collectAsState()
    val peerCount by viewModel.peerCount.collectAsState()
    val powerMode by viewModel.powerModeName.collectAsState()
    val score by viewModel.behaviorScore.collectAsState()
    val rewards by viewModel.rewardStats.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text("SmartLight", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold)

        // Status Card
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(10.dp)
                            .clip(CircleShape)
                            .background(if (isRunning) Color.Green else Color.Red)
                    )
                    Spacer(Modifier.width(8.dp))
                    Text(
                        if (isRunning) "Node Active" else "Node Stopped",
                        style = MaterialTheme.typography.titleMedium
                    )
                    Spacer(Modifier.weight(1f))
                    AssistChip(onClick = {}, label = { Text(powerMode) })
                }
                Spacer(Modifier.height(12.dp))
                Row {
                    Column {
                        Text("Synced Block", style = MaterialTheme.typography.bodySmall, color = Color.Gray)
                        Text("#$syncedBlock", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
                    }
                    Spacer(Modifier.weight(1f))
                    Column(horizontalAlignment = Alignment.End) {
                        Text("Peers", style = MaterialTheme.typography.bodySmall, color = Color.Gray)
                        Text("$peerCount", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }

        // Behavior Score Card
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Behavior Score", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(8.dp))
                Row(verticalAlignment = Alignment.Bottom) {
                    Text("${score.total}", fontSize = 36.sp, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.width(4.dp))
                    Text("/ 10000", style = MaterialTheme.typography.bodySmall, color = Color.Gray)
                }
                Spacer(Modifier.height(12.dp))
                ScoreBar("Liveness (30%)", score.liveness, Color(0xFF4CAF50))
                ScoreBar("Correctness (20%)", score.correctness, Color(0xFF2196F3))
                ScoreBar("Cooperation (25%)", score.cooperation, Color(0xFF9C27B0))
                ScoreBar("Consistency (10%)", score.consistency, Color(0xFFFF9800))
                ScoreBar("Signal (15%)", score.signalSovereignty, Color(0xFF00BCD4))
            }
        }

        // Rewards Card
        ElevatedCard(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Rewards", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(8.dp))
                Row {
                    Column {
                        Text("Total Earned", style = MaterialTheme.typography.bodySmall, color = Color.Gray)
                        Text("${rewards.totalFormatted} PROBE", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
                    }
                    Spacer(Modifier.weight(1f))
                    Column(horizontalAlignment = Alignment.End) {
                        Text("This Epoch", style = MaterialTheme.typography.bodySmall, color = Color.Gray)
                        Text("${rewards.epochFormatted} PROBE", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }

        // Action Buttons
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Button(
                onClick = { viewModel.toggleNode() },
                modifier = Modifier.weight(1f),
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (isRunning) Color(0xFFF44336) else Color(0xFF4CAF50)
                )
            ) {
                Text(if (isRunning) "Stop Node" else "Start Node")
            }
            Button(
                onClick = { viewModel.cyclePowerMode() },
                modifier = Modifier.weight(1f)
            ) {
                Text("Power Mode")
            }
        }
    }
}

@Composable
private fun ScoreBar(label: String, value: Long, color: Color) {
    val fraction by animateFloatAsState(targetValue = value / 10000f, label = "score")
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier.padding(vertical = 2.dp)
    ) {
        Text(label, style = MaterialTheme.typography.bodySmall, modifier = Modifier.width(120.dp))
        Box(
            modifier = Modifier
                .weight(1f)
                .height(8.dp)
                .clip(RoundedCornerShape(4.dp))
                .background(Color.Gray.copy(alpha = 0.2f))
        ) {
            Box(
                modifier = Modifier
                    .fillMaxHeight()
                    .fillMaxWidth(fraction)
                    .clip(RoundedCornerShape(4.dp))
                    .background(color)
            )
        }
        Spacer(Modifier.width(8.dp))
        Text("$value", style = MaterialTheme.typography.labelSmall, modifier = Modifier.width(40.dp))
    }
}
