# Keep gomobile generated classes
-keep class gprobe.** { *; }
-keep class go.** { *; }

# Keep Dilithium key store
-keepclassmembers class * {
    @com.google.gson.annotations.SerializedName <fields>;
}
