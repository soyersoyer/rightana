import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router, Params } from '@angular/router';

import { UserComponent } from '../user/user.component';
import { BackendService } from "../backend.service"
import { ToastyService } from '../toasty/toasty.module';

@Component({
  selector: 'rana-verify-email',
  templateUrl: './verify-email.component.html',
})
export class VerifyEmailComponent implements OnInit {

  constructor(
    private user: UserComponent,
    private backend: BackendService,
    private toasty: ToastyService,
    private route: ActivatedRoute,
    private router: Router,
  ) { }

  ngOnInit() {
    this.route.queryParams.forEach((params: Params) => {
      this.backend.verifyEmail(this.user.user, params['verification_key']).subscribe(_ => {
        this.toasty.success('Email verification complete!')
        this.router.navigateByUrl('/settings/profile');
      });
    });
  }

}
