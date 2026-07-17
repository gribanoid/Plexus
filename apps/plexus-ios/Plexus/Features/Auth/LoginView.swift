import SwiftUI

struct LoginView: View {
    @EnvironmentObject var authStore: AuthStore
    @State private var email = ""
    @State private var password = ""
    @State private var showRegister = false

    private let brand = Color(hex: "#0052CC")

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                VStack(spacing: 8) {
                    RoundedRectangle(cornerRadius: 16)
                        .fill(brand)
                        .frame(width: 64, height: 64)
                        .overlay(
                            Text("P")
                                .font(.system(size: 32, weight: .bold))
                                .foregroundStyle(.white)
                        )
                    Text("Plexus")
                        .font(.largeTitle.bold())
                    Text("Sign in to your workspace")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                VStack(spacing: 12) {
                    TextField("Email or username", text: $email)
                        .textContentType(.username)
                        .keyboardType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .textFieldStyle(.roundedBorder)

                    SecureField("Password", text: $password)
                        .textContentType(.password)
                        .textFieldStyle(.roundedBorder)

                    if let error = authStore.errorMessage {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    Button {
                        Task { await authStore.login(email: email, password: password) }
                    } label: {
                        HStack {
                            if authStore.isLoading {
                                ProgressView().tint(.white)
                            }
                            Text(authStore.isLoading ? "Signing in…" : "Sign in")
                                .font(.headline)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(brand)
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                        .opacity(authStore.isLoading || email.isEmpty || password.isEmpty ? 0.5 : 1)
                    }
                    .buttonStyle(.plain)
                    .disabled(authStore.isLoading || email.isEmpty || password.isEmpty)
                }

                Button("Don't have an account? Create one") {
                    showRegister = true
                }
                .font(.subheadline)
                .foregroundStyle(.secondary)

                Spacer()
            }
            .padding(.horizontal, 24)
            .navigationDestination(isPresented: $showRegister) {
                RegisterView()
            }
            .onAppear {
                authStore.clearError()
            }
        }
    }
}
