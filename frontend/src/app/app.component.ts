import { Component } from '@angular/core';

import { AuthService } from './backend.service';

@Component({
  selector: 'rana-root',
  templateUrl: './app.component.html',
})
export class AppComponent {

  constructor(
    private auth: AuthService,
  ) { }

  get loggedIn(): boolean {
    return this.auth.loggedIn;
  }

  get user(): string {
    return this.auth.user;
  }

  get isAdmin(): boolean {
    return this.auth.isAdmin;
  }

}
