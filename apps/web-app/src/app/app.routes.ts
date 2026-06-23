import { Routes } from '@angular/router'

export const routes: Routes = [{
    path: "/app", 
    loadChildren: () => import("./views/app/dashboard/dashboard").then(m => m.Dashboard),
    canActivate: []
}]
