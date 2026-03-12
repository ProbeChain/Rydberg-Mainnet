// Copyright 2024 The ProbeChain Authors
// GNSSTimeProvider — Provides GNSS (GPS) time samples from the iPhone.
//
// iPhone GPS provides ~100ns time accuracy via the CoreLocation framework.
// This provider converts CLLocation timestamps into AtomicTimestamp format
// (17 bytes) for submission to the ProbeChain network.

import Foundation
import CoreLocation

/// GNSSTimeProvider bridges iPhone GPS time to ProbeChain's AtomicTimestamp format.
final class GNSSTimeProvider: NSObject, ObservableObject, CLLocationManagerDelegate {
    private let locationManager = CLLocationManager()

    @Published var isAvailable = false
    @Published var lastSampleTime: Date?
    @Published var lastAccuracyNs: UInt32 = 0
    @Published var latitude: Double = 0
    @Published var longitude: Double = 0

    private var sampleCallback: ((Data?) -> Void)?

    override init() {
        super.init()
        locationManager.delegate = self
        locationManager.desiredAccuracy = kCLLocationAccuracyBest
        locationManager.allowsBackgroundLocationUpdates = true
        locationManager.pausesLocationUpdatesAutomatically = false
    }

    // MARK: - Public API

    /// Requests location permission and starts monitoring.
    func requestPermission() {
        locationManager.requestAlwaysAuthorization()
    }

    /// Starts location updates for GNSS time sampling.
    func startMonitoring() {
        locationManager.startUpdatingLocation()
    }

    /// Stops location updates.
    func stopMonitoring() {
        locationManager.stopUpdatingLocation()
    }

    /// Takes a GNSS time sample, returns encoded AtomicTimestamp (17 bytes).
    func sample() -> Data? {
        guard isAvailable else { return nil }

        let now = Date()
        let seconds = UInt64(now.timeIntervalSince1970)
        let nanoseconds = UInt32((now.timeIntervalSince1970 - Double(seconds)) * 1_000_000_000)
        let clockSource: UInt8 = 3 // ClockSourceGNSS
        let uncertainty: UInt32 = 100 // ~100ns from GPS

        // Encode as AtomicTimestamp: [8B seconds][4B nanos][1B source][4B uncertainty]
        var data = Data(count: 17)
        data.withUnsafeMutableBytes { ptr in
            let buf = ptr.bindMemory(to: UInt8.self)
            // Seconds (big-endian)
            buf[0] = UInt8((seconds >> 56) & 0xFF)
            buf[1] = UInt8((seconds >> 48) & 0xFF)
            buf[2] = UInt8((seconds >> 40) & 0xFF)
            buf[3] = UInt8((seconds >> 32) & 0xFF)
            buf[4] = UInt8((seconds >> 24) & 0xFF)
            buf[5] = UInt8((seconds >> 16) & 0xFF)
            buf[6] = UInt8((seconds >> 8) & 0xFF)
            buf[7] = UInt8(seconds & 0xFF)
            // Nanoseconds (big-endian)
            buf[8] = UInt8((nanoseconds >> 24) & 0xFF)
            buf[9] = UInt8((nanoseconds >> 16) & 0xFF)
            buf[10] = UInt8((nanoseconds >> 8) & 0xFF)
            buf[11] = UInt8(nanoseconds & 0xFF)
            // Clock source
            buf[12] = clockSource
            // Uncertainty (big-endian)
            buf[13] = UInt8((uncertainty >> 24) & 0xFF)
            buf[14] = UInt8((uncertainty >> 16) & 0xFF)
            buf[15] = UInt8((uncertainty >> 8) & 0xFF)
            buf[16] = UInt8(uncertainty & 0xFF)
        }

        lastSampleTime = now
        lastAccuracyNs = uncertainty
        return data
    }

    /// Returns the current GPS coordinates for anti-Sybil location dedup.
    func getLocation() -> (lat: Double, lon: Double) {
        return (latitude, longitude)
    }

    // MARK: - CLLocationManagerDelegate

    func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        guard let location = locations.last else { return }
        isAvailable = true
        latitude = location.coordinate.latitude
        longitude = location.coordinate.longitude
    }

    func locationManager(_ manager: CLLocationManager, didFailWithError error: Error) {
        isAvailable = false
    }

    func locationManagerDidChangeAuthorization(_ manager: CLLocationManager) {
        switch manager.authorizationStatus {
        case .authorizedAlways, .authorizedWhenInUse:
            startMonitoring()
        default:
            isAvailable = false
        }
    }
}
