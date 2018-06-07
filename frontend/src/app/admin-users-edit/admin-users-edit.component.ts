import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ToastyService } from 'ng2-toasty';

import { UserInfo, UserUpdate, BackendService } from '../backend.service';
import { RValidators } from '../forms/rvalidators';

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
    private router: Router,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, RValidators.userName],
      email: [null, RValidators.email],
      password: [null, RValidators.password],
      is_admin: [null, [Validators.required]],
      disable_pw_change: [null, [Validators.required]],
      limit_collections: [null, [Validators.required]],
      collection_limit: [null, [Validators.required]],
      disable_user_deletion: [null, [Validators.required]],
    });
    this.route.params.forEach((params: Params) => {
      this.getUser(params['name']);
    });
  }

  getUser(name: string) {
    this.backend.getUserInfo(name)
      .subscribe(user => {
        this.user = user;
        this.form.patchValue(user);
      });
  }

  update() {
    this.backend.updateUser(this.user.name, this.form.value)
    .subscribe(_ => {
      this.toasty.success("Update success");
      this.router.navigate(["..", this.form.value.name], {relativeTo: this.route});
    });
  }
}
