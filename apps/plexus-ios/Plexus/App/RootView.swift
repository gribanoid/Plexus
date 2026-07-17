import SwiftUI

struct RootView: View {
    @EnvironmentObject var authStore: AuthStore

    var body: some View {
        Group {
            if authStore.isAuthenticated {
                MainTabView()
            } else {
                LoginView()
            }
        }
        .animation(.easeInOut, value: authStore.isAuthenticated)
        .task {
            if authStore.isAuthenticated {
                await authStore.fetchMe()
            }
        }
    }
}
