import SwiftUI

struct CreateWorkspaceSheet: View {
    var onCreated: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var slug = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private let api = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Workspace name", text: $name)
                    TextField("URL slug (optional)", text: $slug)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                } footer: {
                    Text("A workspace is where your team manages projects.")
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("Create workspace")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSubmitting ? "Creating…" : "Create") {
                        Task { await create() }
                    }
                    .disabled(isSubmitting || name.trimmingCharacters(in: .whitespaces).count < 2)
                }
            }
        }
    }

    private func create() async {
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }

        do {
            struct Body: Encodable {
                let name: String
                let slug: String?
            }
            let trimmedSlug = slug.trimmingCharacters(in: .whitespaces)
            let _: Organization = try await api.request(
                "orgs",
                method: "POST",
                body: Body(
                    name: name.trimmingCharacters(in: .whitespaces),
                    slug: trimmedSlug.isEmpty ? nil : trimmedSlug
                )
            )
            onCreated()
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
