import { Component, signal } from '@angular/core'
import { RouterOutlet, RouterLinkWithHref } from '@angular/router'
import { HlmButton } from '@dots-beacon/ui/button'

@Component({
    selector: 'app-root',
    imports: [RouterOutlet, HlmButton, RouterLinkWithHref],
    templateUrl: './app.html',
    styleUrl: './app.css',
})
export class App {
}
