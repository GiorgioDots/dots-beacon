import { Injectable, signal } from '@angular/core'
import Keycloak from 'keycloak-js'
import { environment } from '../../../environments/environment'

@Injectable({ providedIn: 'root' })
export class KeycloakService {
    private readonly keycloak = new Keycloak({
        url: environment.keycloak.url,
        realm: environment.keycloak.realm,
        clientId: environment.keycloak.clientId,
    })

    private readonly _isAuthenticated = signal(false)
    readonly isAuthenticated = this._isAuthenticated.asReadonly()

    async initialize(): Promise<void> {
        // await this.keycloak.logout();
        const authenticated = await this.keycloak.init({
            onLoad: 'check-sso',
            silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
        })
        this._isAuthenticated.set(authenticated)

        this.keycloak.onTokenExpired = () => {
            this.keycloak.updateToken(30).catch(() => this._isAuthenticated.set(false))
        }
    }

    login(): Promise<void> {
        return this.keycloak.login()
    }

    logout(): Promise<void> {
        return this.keycloak.logout({ redirectUri: window.location.origin })
    }

    getToken(): string | undefined {
        return this.keycloak.token
    }
}
