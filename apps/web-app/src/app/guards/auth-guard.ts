import { inject } from '@angular/core'
import { CanActivateFn } from '@angular/router'
import { KeycloakService } from '../services/auth/keycloak.service'

export const authGuard: CanActivateFn = () => {
    const keycloak = inject(KeycloakService)

    if (keycloak.isAuthenticated()) {
        return true
    }

    keycloak.login()
    return false
}
