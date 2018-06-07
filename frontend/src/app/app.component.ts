import { Component } from '@angular/core';
import { Router, RouterEvent, NavigationEnd } from '@angular/router';

import { AuthService, BackendService } from './backend.service';

declare var rightana: any;

@Component({
  selector: 'rana-root',
  templateUrl: './app.component.html',
})
export class AppComponent {

  constructor(
    private auth: AuthService,
    private backend: BackendService,
    private router: Router,
  ) {
    this.setupRightana();
  }

  setupRightana() {
    this.backend.getConfig().subscribe(config => {
      if (config.tracking_id) {
        rightana('setup', '/api', config.tracking_id);
        rightana('trackPageview');
        this.router.events.subscribe((event: RouterEvent) => {
          if (event instanceof NavigationEnd) {
            rightana('trackPageview');
          }
        });
      }
    });
  }

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
