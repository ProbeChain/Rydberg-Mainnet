// Copyright 2024 The ProbeChain Authors

import Foundation
import Combine

/// ScoreService tracks the SmartLight node's behavior score.
final class ScoreService: ObservableObject {
    static let shared = ScoreService()

    @Published var score = BehaviorScoreData()

    private let bridge = GoNodeBridge()
    private var timer: Timer?

    private init() {}

    func startTracking() {
        timer = Timer.scheduledTimer(withTimeInterval: 10.0, repeats: true) { [weak self] _ in
            self?.refreshScore()
        }
    }

    func stopTracking() {
        timer?.invalidate()
        timer = nil
    }

    func refreshScore() {
        let json = bridge.getBehaviorScore()
        guard let data = json.data(using: .utf8) else { return }

        do {
            let decoded = try JSONDecoder().decode(SmartLightScoreJSON.self, from: data)
            DispatchQueue.main.async {
                self.score = BehaviorScoreData(
                    total: decoded.total,
                    liveness: decoded.liveness,
                    correctness: decoded.correctness,
                    cooperation: decoded.cooperation,
                    consistency: decoded.consistency,
                    signalSovereignty: decoded.signalSovereignty
                )
            }
        } catch {
            // Score not available yet
        }
    }
}

// JSON mapping for Go SmartLightScore
private struct SmartLightScoreJSON: Decodable {
    let total: UInt64
    let liveness: UInt64
    let correctness: UInt64
    let cooperation: UInt64
    let consistency: UInt64
    let signalSovereignty: UInt64
    let lastUpdate: UInt64
}
