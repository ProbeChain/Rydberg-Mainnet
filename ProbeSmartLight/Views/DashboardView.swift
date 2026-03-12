// Copyright 2024 The ProbeChain Authors

import SwiftUI

struct DashboardView: View {
    @EnvironmentObject var nodeService: NodeService
    @EnvironmentObject var scoreService: ScoreService
    @EnvironmentObject var rewardService: RewardService

    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 16) {
                    // Node Status Card
                    StatusCard(
                        isRunning: nodeService.isRunning,
                        syncedBlock: nodeService.syncedBlock,
                        peerCount: nodeService.peerCount,
                        powerMode: nodeService.powerModeName
                    )

                    // Behavior Score Card
                    ScoreCard(score: scoreService.score)

                    // Reward Stats Card
                    RewardCard(stats: rewardService.stats)

                    // Error display
                    if !nodeService.lastError.isEmpty {
                        Text(nodeService.lastError)
                            .font(.caption)
                            .foregroundColor(.red)
                            .padding()
                            .background(Color.red.opacity(0.1))
                            .cornerRadius(8)
                            .padding(.horizontal)
                    }

                    // Quick Actions
                    HStack(spacing: 12) {
                        ActionButton(
                            title: nodeService.isRunning ? "Stop Node" : "Start Node",
                            icon: nodeService.isRunning ? "stop.fill" : "play.fill",
                            color: nodeService.isRunning ? .red : .green
                        ) {
                            if nodeService.isRunning {
                                nodeService.stopNode()
                            } else {
                                nodeService.startNode()
                            }
                        }

                        ActionButton(
                            title: "Power Mode",
                            icon: "battery.100",
                            color: .blue
                        ) {
                            nodeService.cyclePowerMode()
                        }
                    }
                    .padding(.horizontal)
                }
                .padding(.vertical)
            }
            .navigationTitle("SmartLight")
        }
    }
}

// MARK: - Subviews

struct StatusCard: View {
    let isRunning: Bool
    let syncedBlock: UInt64
    let peerCount: Int
    let powerMode: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Circle()
                    .fill(isRunning ? Color.green : Color.red)
                    .frame(width: 10, height: 10)
                Text(isRunning ? "Node Active" : "Node Stopped")
                    .font(.headline)
                Spacer()
                Text(powerMode)
                    .font(.caption)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.blue.opacity(0.2))
                    .cornerRadius(8)
            }

            HStack {
                VStack(alignment: .leading) {
                    Text("Synced Block")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text("#\(syncedBlock)")
                        .font(.title3)
                        .fontWeight(.semibold)
                }
                Spacer()
                VStack(alignment: .trailing) {
                    Text("Peers")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text("\(peerCount)")
                        .font(.title3)
                        .fontWeight(.semibold)
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(radius: 2)
        .padding(.horizontal)
    }
}

struct ScoreCard: View {
    let score: BehaviorScoreData

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Behavior Score")
                .font(.headline)

            HStack {
                Text("\(score.total)")
                    .font(.system(size: 36, weight: .bold))
                Text("/ 10000")
                    .font(.caption)
                    .foregroundColor(.secondary)
                Spacer()
            }

            // Dimension bars
            ScoreDimension(name: "Liveness", value: score.liveness, weight: 30, color: .green)
            ScoreDimension(name: "Correctness", value: score.correctness, weight: 20, color: .blue)
            ScoreDimension(name: "Cooperation", value: score.cooperation, weight: 25, color: .purple)
            ScoreDimension(name: "Consistency", value: score.consistency, weight: 10, color: .orange)
            ScoreDimension(name: "Signal", value: score.signalSovereignty, weight: 15, color: .cyan)
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(radius: 2)
        .padding(.horizontal)
    }
}

struct ScoreDimension: View {
    let name: String
    let value: UInt64
    let weight: Int
    let color: Color

    var body: some View {
        HStack {
            Text("\(name) (\(weight)%)")
                .font(.caption)
                .frame(width: 120, alignment: .leading)
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    Rectangle()
                        .fill(Color.gray.opacity(0.2))
                        .cornerRadius(4)
                    Rectangle()
                        .fill(color)
                        .cornerRadius(4)
                        .frame(width: geo.size.width * CGFloat(value) / 10000.0)
                }
            }
            .frame(height: 8)
            Text("\(value)")
                .font(.caption2)
                .frame(width: 40, alignment: .trailing)
        }
    }
}

struct RewardCard: View {
    let stats: RewardStatsData

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Rewards")
                .font(.headline)
            HStack {
                VStack(alignment: .leading) {
                    Text("Total Earned")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text("\(stats.totalRewardsFormatted) PROBE")
                        .font(.title3)
                        .fontWeight(.semibold)
                }
                Spacer()
                VStack(alignment: .trailing) {
                    Text("This Epoch")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text("\(stats.epochRewardsFormatted) PROBE")
                        .font(.title3)
                        .fontWeight(.semibold)
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(radius: 2)
        .padding(.horizontal)
    }
}

struct ActionButton: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack {
                Image(systemName: icon)
                Text(title)
            }
            .frame(maxWidth: .infinity)
            .padding()
            .background(color)
            .foregroundColor(.white)
            .cornerRadius(12)
        }
    }
}

// MARK: - Data Models

struct BehaviorScoreData {
    var total: UInt64 = 5000
    var liveness: UInt64 = 10000
    var correctness: UInt64 = 10000
    var cooperation: UInt64 = 10000
    var consistency: UInt64 = 10000
    var signalSovereignty: UInt64 = 5000
}

struct RewardStatsData {
    var totalRewardsFormatted: String = "0.00"
    var epochRewardsFormatted: String = "0.00"
    var currentEpoch: UInt64 = 0
}
