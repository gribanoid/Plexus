import SwiftUI

@main
struct PlexusApp: App {
    @StateObject private var authStore = AuthStore()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environmentObject(authStore)
                .preferredColorScheme(nil)
        }
    }
}
