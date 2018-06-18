import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { FormBuilder, FormGroup } from '@angular/forms';

import { RValidators } from '../forms/rvalidators';
import { BackendService } from '../backend.service';
import { ToastyService } from '../toasty/toasty.module';


@Component({
  selector: 'rana-admin-users-create',
  templateUrl: './admin-users-create.component.html',
})
export class AdminUsersCreateComponent implements OnInit {
  form: FormGroup;

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
    });
  }

  create() {
    this.backend.createUserAdmin(this.form.value)
    .subscribe(_ => {
      this.toasty.success("Create success");
      this.router.navigate([".."], {relativeTo: this.route});
    });
  }

}
