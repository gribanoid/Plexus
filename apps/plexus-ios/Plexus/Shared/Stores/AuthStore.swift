import SwiftUI

@MainActor
final class AuthStore: ObservableObject {
    @Published var currentUser: User?
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published private(set) var isAuthenticated: Bool

    private let api = APIClient.shared
    private let keychain = KeychainStore.shared

    init() {
        isAuthenticated = keychain.accessToken != nil
    }

    func clearError() {
        errorMessage = nil
    }

    func login(email: String, password: String) async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            struct LoginBody: Encodable { let email: String; let password: String }
            let pair: TokenPair = try await api.request(
                "auth/login",
                method: "POST",
                body: LoginBody(email: Self.resolveLoginEmail(email), password: password)
            )
            keychain.accessToken = pair.accessToken
            keychain.refreshToken = pair.refreshToken
            isAuthenticated = true
            await fetchMe()
        } catch APIError.unauthorized {
            errorMessage = "Invalid email or password"
            logout()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func register(email: String, password: String, displayName: String) async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            struct RegisterBody: Encodable {
                let email: String
                let password: String
                let displayName: String
            }
            let pair: TokenPair = try await api.request(
                "auth/register",
                method: "POST",
                body: RegisterBody(email: email, password: password, displayName: displayName)
            )
            keychain.accessToken = pair.accessToken
            keychain.refreshToken = pair.refreshToken
            isAuthenticated = true
            await fetchMe()
        } catch APIError.unauthorized {
            errorMessage = "Registration failed"
            logout()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func logout() {
        keychain.accessToken = nil
        keychain.refreshToken = nil
        currentUser = nil
        isAuthenticated = false
    }

    func fetchMe() async {
        do {
            currentUser = try await api.request("me")
            isAuthenticated = true
        } catch APIError.unauthorized {
            logout()
        } catch {
            // Non-fatal: keep session if token exists
        }
    }

    /// Match web: shorthand `admin` → seed user email.
    static func resolveLoginEmail(_ input: String) -> String {
        let trimmed = input.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.lowercased() == "admin" {
            return "admin@plexus.local"
        }
        return trimmed
    }
}
