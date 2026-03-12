// Copyright 2024 The ProbeChain Authors

import Foundation
import Combine

/// RewardService tracks PROBE reward accumulation for the SmartLight node.
final class RewardService: ObservableObject {
    static let shared = RewardService()

    @Published var stats = RewardStatsData()

    private let bridge = GoNodeBridge()
    private var timer: Timer?

    private init() {}

    func startTracking() {
        timer = Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true) { [weak self] _ in
            self?.refreshStats()
        }
    }

    func stopTracking() {
        timer?.invalidate()
        timer = nil
    }

    func refreshStats() {
        let json = bridge.getRewardStats()
        guard let data = json.data(using: .utf8) else { return }

        do {
            let decoded = try JSONDecoder().decode(RewardStatsJSON.self, from: data)
            DispatchQueue.main.async {
                self.stats = RewardStatsData(
                    totalRewardsFormatted: Self.formatProbe(decoded.totalRewards),
                    epochRewardsFormatted: Self.formatProbe(decoded.epochRewards),
                    currentEpoch: decoded.currentEpoch
                )
            }
        } catch {
            // Stats not available yet
        }
    }

    /// Formats a wei string to PROBE with 4 decimal places.
    static func formatProbe(_ weiString: String) -> String {
        guard let wei = Decimal(string: weiString) else { return "0.0000" }
        let probe = wei / Decimal(sign: .plus, exponent: 18, significand: 1)
        let formatter = NumberFormatter()
        formatter.minimumFractionDigits = 4
        formatter.maximumFractionDigits = 4
        return formatter.string(from: probe as NSDecimalNumber) ?? "0.0000"
    }
}

// JSON mapping for Go RewardStats
private struct RewardStatsJSON: Decodable {
    let totalRewards: String
    let epochRewards: String
    let currentEpoch: UInt64
    let lastRewardBlock: UInt64
}
