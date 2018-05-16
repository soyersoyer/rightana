import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ToastyService } from 'ng2-toasty';

import { UserInfo, UserUpdate, BackendService } from '../backend.service';

@Component({
  selector: 'rana-admin-users-edit',
  templateUrl: './admin-users-edit.component.html',
})
export class AdminUsersEditComponent implements OnInit {
  form: FormGroup;
  user: UserInfo;

  constructor(
  	private fb: FormBuilder,
    private backend: BackendService,
    private route: ActivatedRoute,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, [Validators.required, Validators.pattern("^[a-z0-9.]+$")]],
      password: [null, [Validators.required]],
      is_admin: [null, [Validators.required]],
    });
    this.route.params.forEach((params: Params) => {
      this.getUser(params['email']);
    });
  }

  getUser(email: string) {
    this.backend.getUserInfo(email)
      .subscribe(user => {
        this.user = user;
        this.form.patchValue(user);
      });
  }

  update() {
    this.backend.updateUser(this.user.email, this.form.value)
    .subscribe(_ => {
      this.toasty.success("Update success");
    });
  }
}
