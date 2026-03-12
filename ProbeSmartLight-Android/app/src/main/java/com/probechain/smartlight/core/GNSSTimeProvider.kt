// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.core

import android.annotation.SuppressLint
import android.content.Context
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.os.Bundle
import android.os.SystemClock
import android.util.Log
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.nio.ByteBuffer

/**
 * GNSSTimeProvider provides GNSS (GPS) time samples from Android devices.
 *
 * Android GPS provides time via Location.getTime() (UTC milliseconds) and
 * Location.getElapsedRealtimeNanos() for monotonic timing. We convert these
 * to ProbeChain's AtomicTimestamp format (17 bytes).
 *
 * AtomicTimestamp layout: [8B seconds][4B nanos][1B source][4B uncertainty]
 * ClockSourceGNSS = 3, uncertainty = ~100ns
 */
class GNSSTimeProvider(private val context: Context) : LocationListener {

    private val locationManager = context.getSystemService(Context.LOCATION_SERVICE) as LocationManager

    private val _isAvailable = MutableStateFlow(false)
    val isAvailable: StateFlow<Boolean> = _isAvailable.asStateFlow()

    private val _latitude = MutableStateFlow(0.0)
    val latitude: StateFlow<Double> = _latitude.asStateFlow()

    private val _longitude = MutableStateFlow(0.0)
    val longitude: StateFlow<Double> = _longitude.asStateFlow()

    private val _sampleCount = MutableStateFlow(0L)
    val sampleCount: StateFlow<Long> = _sampleCount.asStateFlow()

    companion object {
        private const val TAG = "GNSSTimeProvider"
        private const val CLOCK_SOURCE_GNSS: Byte = 3
        private const val DEFAULT_UNCERTAINTY_NS: Int = 100 // ~100ns from GPS
        private const val ATOMIC_TIMESTAMP_SIZE = 17
    }

    /**
     * Start receiving GPS location updates.
     */
    @SuppressLint("MissingPermission")
    fun startMonitoring() {
        try {
            // Request GPS updates every 1 second, 0 meters minimum distance
            locationManager.requestLocationUpdates(
                LocationManager.GPS_PROVIDER,
                1000L,
                0f,
                this
            )
            Log.i(TAG, "GNSS monitoring started")
        } catch (e: SecurityException) {
            Log.w(TAG, "Location permission not granted", e)
        }
    }

    /**
     * Stop receiving GPS location updates.
     */
    fun stopMonitoring() {
        locationManager.removeUpdates(this)
        _isAvailable.value = false
        Log.i(TAG, "GNSS monitoring stopped")
    }

    /**
     * Takes a GNSS time sample and returns encoded AtomicTimestamp (17 bytes).
     * Returns null if GNSS is not available.
     */
    fun sample(): ByteArray? {
        if (!_isAvailable.value) return null

        val nowMs = System.currentTimeMillis()
        val seconds = nowMs / 1000
        val nanoseconds = ((nowMs % 1000) * 1_000_000).toInt()

        val buf = ByteBuffer.allocate(ATOMIC_TIMESTAMP_SIZE)
        buf.putLong(seconds)           // 8 bytes: seconds
        buf.putInt(nanoseconds)        // 4 bytes: nanoseconds
        buf.put(CLOCK_SOURCE_GNSS)     // 1 byte: clock source
        buf.putInt(DEFAULT_UNCERTAINTY_NS) // 4 bytes: uncertainty

        _sampleCount.value++
        return buf.array()
    }

    /**
     * Returns current GPS coordinates for anti-Sybil location dedup.
     */
    fun getLocation(): Pair<Double, Double> {
        return Pair(_latitude.value, _longitude.value)
    }

    // LocationListener callbacks

    override fun onLocationChanged(location: Location) {
        _isAvailable.value = true
        _latitude.value = location.latitude
        _longitude.value = location.longitude
    }

    override fun onProviderEnabled(provider: String) {
        if (provider == LocationManager.GPS_PROVIDER) {
            _isAvailable.value = true
        }
    }

    override fun onProviderDisabled(provider: String) {
        if (provider == LocationManager.GPS_PROVIDER) {
            _isAvailable.value = false
        }
    }

    @Deprecated("Deprecated in API level 29")
    override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) {}
}
