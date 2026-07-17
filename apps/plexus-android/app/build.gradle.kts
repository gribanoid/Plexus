plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.hilt)
    alias(libs.plugins.ksp)
}

android {
    namespace = "com.plexus"
    compileSdk = 35

    defaultConfig {
        applicationId = "app.plexus"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "1.0.0"

        buildConfigField("String", "API_BASE_URL", "\"http://10.0.2.2:8080/api/v1\"")
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            buildConfigField("String", "API_BASE_URL", "\"https://api.plexus.app/api/v1\"")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }
}

dependencies {
    // Compose
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.tooling.preview)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons)
    implementation(libs.compose.activity)
    debugImplementation(libs.compose.ui.tooling)

    // Navigation
    implementation(libs.navigation.compose)
    implementation(libs.hilt.navigation.compose)

    // Lifecycle
    implementation(libs.lifecycle.viewmodel.compose)
    implementation(libs.lifecycle.runtime.compose)

    // Hilt
    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)

    // Networking
    implementation(libs.retrofit)
    implementation(libs.retrofit.moshi)
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging)
    implementation(libs.moshi.kotlin)
    ksp(libs.moshi.codegen)

    // Coroutines
    implementation(libs.coroutines.android)

    // Storage
    implementation(libs.datastore.preferences)
    implementation(libs.security.crypto)

    // Images
    implementation(libs.coil.compose)

    // DateTime
    implementation(libs.kotlinx.datetime)
}
