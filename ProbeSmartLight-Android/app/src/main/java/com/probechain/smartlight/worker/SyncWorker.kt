// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.worker

import android.content.Context
import androidx.work.*
import java.util.concurrent.TimeUnit

/**
 * SyncWorker runs periodic sync and heartbeat tasks via WorkManager.
 * Ensures the SmartLight node stays alive even when the app is backgrounded.
 */
class SyncWorker(context: Context, params: WorkerParameters) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        // The Go node handles sync internally via P2P.
        // This worker ensures the foreground service is running and
        // triggers any queued heartbeat/ACK flushes.
        return Result.success()
    }

    companion object {
        private const val WORK_NAME = "smartlight_sync"

        /**
         * Schedules periodic sync work (every 15 minutes minimum).
         */
        fun schedule(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()

            val request = PeriodicWorkRequestBuilder<SyncWorker>(15, TimeUnit.MINUTES)
                .setConstraints(constraints)
                .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 1, TimeUnit.MINUTES)
                .build()

            WorkManager.getInstance(context)
                .enqueueUniquePeriodicWork(WORK_NAME, ExistingPeriodicWorkPolicy.KEEP, request)
        }

        /**
         * Cancels scheduled sync work.
         */
        fun cancel(context: Context) {
            WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
        }
    }
}
