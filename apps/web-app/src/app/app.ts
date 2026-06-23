import { Component, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { HlmButton } from "@dots-beacon/ui/button";

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, HlmButton],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  protected readonly title = signal('web-app');
}
