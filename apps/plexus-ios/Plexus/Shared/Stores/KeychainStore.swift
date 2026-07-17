import Foundation
import Security

/// Stores sensitive tokens in the iOS Keychain.
final class KeychainStore {
    static let shared = KeychainStore()
    private init() {}

    private let service = "app.plexus"

    var accessToken: String? {
        get { read(key: "access_token") }
        set { newValue == nil ? delete(key: "access_token") : write(key: "access_token", value: newValue!) }
    }

    var refreshToken: String? {
        get { read(key: "refresh_token") }
        set { newValue == nil ? delete(key: "refresh_token") : write(key: "refresh_token", value: newValue!) }
    }

    // MARK: - Private helpers

    private func write(key: String, value: String) {
        let data = Data(value.utf8)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
            kSecValueData: data,
        ]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    private func read(key: String) -> String? {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne,
        ]
        var result: AnyObject?
        SecItemCopyMatching(query as CFDictionary, &result)
        guard let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    private func delete(key: String) {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key,
        ]
        SecItemDelete(query as CFDictionary)
    }
}
