import SwiftUI

struct CreateProjectSheet: View {
    let orgSlug: String
    var onCreated: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var key = ""
    @State private var description = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private let api = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Project name", text: $name)
                        .onChange(of: name) { _, newValue in
                            if newValue.trimmingCharacters(in: .whitespaces).count >= 2 {
                                key = suggestKey(from: newValue)
                            }
                        }
                    TextField("Key", text: $key)
                        .textInputAutocapitalization(.characters)
                        .autocorrectionDisabled()
                    TextField("Description (optional)", text: $description, axis: .vertical)
                        .lineLimit(2...4)
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("Create project")
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
                let key: String?
                let description: String?
            }
            let trimmedKey = key.trimmingCharacters(in: .whitespaces).uppercased()
            let trimmedDescription = description.trimmingCharacters(in: .whitespaces)
            let _: Project = try await api.request(
                "orgs/\(orgSlug)/projects",
                method: "POST",
                body: Body(
                    name: name.trimmingCharacters(in: .whitespaces),
                    key: trimmedKey.isEmpty ? nil : trimmedKey,
                    description: trimmedDescription.isEmpty ? nil : trimmedDescription
                )
            )
            onCreated()
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func suggestKey(from name: String) -> String {
        let words = name.trimmingCharacters(in: .whitespaces).split(separator: " ")
        var result = ""
        for word in words {
            if result.count >= 4 { break }
            if let ch = word.first(where: { $0.isLetter || $0.isNumber }) {
                result.append(ch.uppercased())
            }
        }
        return result.isEmpty ? "PRJ" : result
    }
}
