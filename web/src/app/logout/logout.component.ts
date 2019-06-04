import { Component, OnInit } from '@angular/core';
import { AuthService, BackendService } from '../backend.service';
import { Router } from '@angular/router';

import { ToastyService } from '../toasty/toasty.module';

@Component({
  selector: 'rana-logout',
  templateUrl: './logout.component.html',
})
export class LogoutComponent implements OnInit {

  constructor(
    private auth: AuthService,
    private backend: BackendService,
    private toasty: ToastyService,
    private router: Router,
  ) { }

  ngOnInit() {
    this.backend
      .deleteAuthToken(this.auth.token)
      .subscribe(
        () => {
          this.toasty.success('Logout success');
          this.auth.unset();
          this.router.navigate(['/']);
        },
        () => {
          this.auth.unset();
          this.router.navigate(['/']);
        },
      );
  }

}
