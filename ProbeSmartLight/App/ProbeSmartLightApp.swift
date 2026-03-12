// Copyright 2024 The ProbeChain Authors
// ProbeSmartLight — Turn second-hand iPhones into SmartLight nodes.

import SwiftUI
import BackgroundTasks

@main
struct ProbeSmartLightApp: App {
    @StateObject private var nodeService = NodeService.shared
    @StateObject private var scoreService = ScoreService.shared
    @StateObject private var rewardService = RewardService.shared

    init() {
        registerBackgroundTasks()
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(nodeService)
                .environmentObject(scoreService)
                .environmentObject(rewardService)
        }
    }

    // MARK: - Background Tasks

    private func registerBackgroundTasks() {
        // Register background refresh for heartbeat and sync
        BGTaskScheduler.shared.register(
            forTaskWithIdentifier: "com.probechain.smartlight.sync",
            using: nil
        ) { task in
            self.handleBackgroundSync(task: task as! BGAppRefreshTask)
        }

        // Register background processing for ACK batches
        BGTaskScheduler.shared.register(
            forTaskWithIdentifier: "com.probechain.smartlight.ack",
            using: nil
        ) { task in
            self.handleBackgroundAck(task: task as! BGProcessingTask)
        }
    }

    private func handleBackgroundSync(task: BGAppRefreshTask) {
        // Schedule next refresh
        scheduleBackgroundSync()

        task.expirationHandler = {
            // Clean up if needed
        }

        // The Go node continues syncing in background
        // Just send a heartbeat if due
        Task {
            await nodeService.sendHeartbeatIfDue()
            task.setTaskCompleted(success: true)
        }
    }

    private func handleBackgroundAck(task: BGProcessingTask) {
        task.expirationHandler = {
            // Flush any pending ACKs
        }

        Task {
            await nodeService.flushPendingAcks()
            task.setTaskCompleted(success: true)
        }
    }

    private func scheduleBackgroundSync() {
        let request = BGAppRefreshTaskRequest(
            identifier: "com.probechain.smartlight.sync"
        )
        request.earliestBeginDate = Date(timeIntervalSinceNow: 60) // 1 minute
        try? BGTaskScheduler.shared.submit(request)
    }
}
