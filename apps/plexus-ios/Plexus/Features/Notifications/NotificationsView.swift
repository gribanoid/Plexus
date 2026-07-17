import SwiftUI

@MainActor
final class NotificationsViewModel: ObservableObject {
    @Published var notifications: [PlexusNotification] = []
    @Published var isLoading = false
    @Published var isMarkingRead = false

    private let api = APIClient.shared

    func load() async {
        isLoading = notifications.isEmpty
        defer { isLoading = false }
        do {
            let res: ListResponse<PlexusNotification> = try await api.request("notifications")
            notifications = res.items
        } catch {}
    }

    func markRead(id: String) async {
        guard let index = notifications.firstIndex(where: { $0.id == id }),
              !notifications[index].read else { return }
        isMarkingRead = true
        defer { isMarkingRead = false }
        do {
            try await api.requestEmpty("notifications/\(id)/read", method: "POST")
            await load()
        } catch {}
    }

    func markAllRead() async {
        isMarkingRead = true
        defer { isMarkingRead = false }
        do {
            try await api.requestEmpty("notifications/read-all", method: "POST")
            await load()
        } catch {}
    }
}

struct NotificationsView: View {
    @StateObject private var vm = NotificationsViewModel()

    var body: some View {
        Group {
            if vm.isLoading && vm.notifications.isEmpty {
                ProgressView()
            } else if vm.notifications.isEmpty {
                ContentUnavailableView("No notifications", systemImage: "bell.slash")
            } else {
                List(vm.notifications) { n in
                    Button {
                        Task { await vm.markRead(id: n.id) }
                    } label: {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(n.title).font(.subheadline.weight(.medium))
                            if let body = n.body {
                                Text(body).font(.caption).foregroundStyle(.secondary)
                            }
                            Text(n.createdAt, style: .relative)
                                .font(.caption2).foregroundStyle(.tertiary)
                        }
                        .opacity(n.read ? 0.6 : 1)
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        .navigationTitle("Notifications")
        .toolbar {
            if vm.notifications.contains(where: { !$0.read }) {
                ToolbarItem(placement: .primaryAction) {
                    Button("Mark all read") {
                        Task { await vm.markAllRead() }
                    }
                    .disabled(vm.isMarkingRead)
                }
            }
        }
        .task { await vm.load() }
        .refreshable { await vm.load() }
    }
}
