import SwiftUI

struct MainTabView: View {
    var body: some View {
        TabView {
            NavigationStack {
                OrgsListView()
            }
            .tabItem {
                Label("Projects", systemImage: "folder")
            }

            NavigationStack {
                IssuesListView()
            }
            .tabItem {
                Label("Issues", systemImage: "checklist")
            }

            NavigationStack {
                NotificationsView()
            }
            .tabItem {
                Label("Notifications", systemImage: "bell")
            }

            NavigationStack {
                AccountView()
            }
            .tabItem {
                Label("Account", systemImage: "person.circle")
            }
        }
        .tint(Color(hex: "#0052CC"))
    }
}
