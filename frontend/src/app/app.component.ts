import { Component } from '@angular/core';

import { AuthService } from './backend.service';

@Component({
  selector: 'k20a-root',
  templateUrl: './app.component.html',
})
export class AppComponent {

  constructor(
    private auth: AuthService,
  ) { }

  get loggedIn(): boolean {
    return this.auth.loggedIn;
  }

}
