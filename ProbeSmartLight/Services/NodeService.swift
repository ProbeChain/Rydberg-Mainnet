// Copyright 2024 The ProbeChain Authors

import Foundation
import Combine
import UIKit

/// NodeService manages the SmartLight node lifecycle and monitoring.
final class NodeService: ObservableObject {
    static let shared = NodeService()

    @Published var isRunning = false
    @Published var syncedBlock: UInt64 = 0
    @Published var peerCount: Int = 0
    @Published var powerModeName: String = "Full"

    private let bridge = GoNodeBridge()
    private var timer: Timer?
    private var currentPowerMode: Int = 0

    private init() {
        // Monitor battery state for auto power mode switching
        UIDevice.current.isBatteryMonitoringEnabled = true
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(batteryStateChanged),
            name: UIDevice.batteryStateDidChangeNotification,
            object: nil
        )
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(batteryLevelChanged),
            name: UIDevice.batteryLevelDidChangeNotification,
            object: nil
        )
    }

    // MARK: - Node Lifecycle

    @Published var lastError: String = ""

    func startNode() {
        let dataDir = getDataDir()
        print("[ProbeSmartLight] Starting node, dataDir: \(dataDir)")
        do {
            try bridge.start(dataDir: dataDir)
            isRunning = true
            lastError = ""
            startMonitoring()
            updatePowerModeFromBattery()
            print("[ProbeSmartLight] Node started successfully")
        } catch {
            lastError = error.localizedDescription
            print("[ProbeSmartLight] FAILED to start node: \(error)")
        }
    }

    func stopNode() {
        do {
            try bridge.stop()
            isRunning = false
            stopMonitoring()
        } catch {
            print("Failed to stop SmartLight node: \(error)")
        }
    }

    // MARK: - Power Mode

    func setPowerMode(_ mode: Int) {
        currentPowerMode = mode
        bridge.setPowerMode(mode)
        switch mode {
        case 0: powerModeName = "Full"
        case 1: powerModeName = "Eco"
        case 2: powerModeName = "Sleep"
        default: powerModeName = "Unknown"
        }
    }

    func cyclePowerMode() {
        let next = (currentPowerMode + 1) % 3
        setPowerMode(next)
    }

    // MARK: - Background Tasks

    func sendHeartbeatIfDue() async {
        // Heartbeat is handled by the Go engine automatically
    }

    func flushPendingAcks() async {
        // ACK flushing is handled by the Go engine automatically
    }

    // MARK: - Monitoring

    private func startMonitoring() {
        timer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { [weak self] _ in
            self?.updateStats()
        }
    }

    private func stopMonitoring() {
        timer?.invalidate()
        timer = nil
    }

    private func updateStats() {
        peerCount = bridge.getPeerCount()
        syncedBlock = bridge.getSyncedBlockNumber()
    }

    // MARK: - Battery-Adaptive Power Mode

    @objc private func batteryStateChanged() {
        updatePowerModeFromBattery()
    }

    @objc private func batteryLevelChanged() {
        updatePowerModeFromBattery()
    }

    private func updatePowerModeFromBattery() {
        let state = UIDevice.current.batteryState
        let level = UIDevice.current.batteryLevel

        if state == .charging || state == .full {
            setPowerMode(0) // Full
        } else if level > 0.30 {
            setPowerMode(1) // Eco
        } else if level > 0 {
            setPowerMode(2) // Sleep
        }
    }

    // MARK: - Helpers

    private func getDataDir() -> String {
        let docs = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        let nodeDir = docs.appendingPathComponent("smartlight-node")
        try? FileManager.default.createDirectory(at: nodeDir, withIntermediateDirectories: true)
        return nodeDir.path
    }
}
