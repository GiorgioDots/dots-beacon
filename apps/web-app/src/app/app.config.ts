import {
    ApplicationConfig,
    inject,
    provideBrowserGlobalErrorListeners,
    provideAppInitializer,
} from '@angular/core'
import { provideRouter } from '@angular/router'
import { KeycloakService } from './services/auth/keycloak.service'
import { routes } from './app.routes'

export const appConfig: ApplicationConfig = {
    providers: [
        provideBrowserGlobalErrorListeners(),
        provideRouter(routes),
        provideAppInitializer(() => {
            const keycloak = inject(KeycloakService)
            return keycloak.initialize()
        }),
    ],
}
