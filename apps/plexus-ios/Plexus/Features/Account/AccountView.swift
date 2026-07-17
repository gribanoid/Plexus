import SwiftUI

struct AccountView: View {
    @EnvironmentObject var authStore: AuthStore

    var body: some View {
        List {
            if let user = authStore.currentUser {
                Section {
                    HStack(spacing: 12) {
                        Circle()
                            .fill(Color(hex: "#0052CC"))
                            .frame(width: 48, height: 48)
                            .overlay(
                                Text(user.displayName.prefix(1).uppercased())
                                    .font(.title2.bold())
                                    .foregroundStyle(.white)
                            )
                        VStack(alignment: .leading) {
                            Text(user.displayName).font(.headline)
                            Text(user.email).font(.subheadline).foregroundStyle(.secondary)
                        }
                    }
                    .padding(.vertical, 4)
                }
            }

            Section {
                Button("Sign out", role: .destructive) {
                    authStore.logout()
                }
            }
        }
        .navigationTitle("Account")
    }
}
