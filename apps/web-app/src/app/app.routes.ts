import { Routes } from '@angular/router'
import { authGuard } from './guards/auth-guard'

export const routes: Routes = [
    {
        path: 'app',
        loadComponent: () => import('./views/app/dashboard/dashboard').then(m => m.Dashboard),
        canActivate: [authGuard],
    },
]
