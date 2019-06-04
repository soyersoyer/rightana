import { Component, OnInit } from '@angular/core';

import { AuthService } from '../backend.service';
import { Router } from '@angular/router';

@Component({
  selector: 'rana-settings',
  templateUrl: './settings.component.html',
})
export class SettingsComponent implements OnInit {

  constructor(
    private auth: AuthService,
    private router: Router,
  ) {
    if (!this.auth.loggedIn) {
      this.router.navigateByUrl('/login');
    }
  }

  ngOnInit() {
  }

}
