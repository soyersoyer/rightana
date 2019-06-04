import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';

import { UserComponent } from '../user/user.component';
import { BackendService } from '../backend.service';
import { RValidators } from '../forms/rvalidators';

@Component({
  selector: 'rana-collection-create',
  templateUrl: './create.component.html',
})
export class CollectionCreateComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private backend: BackendService,
    private router: Router,
    private route: ActivatedRoute,
    private user: UserComponent,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, RValidators.collectionName],
    });
  }

  create() {
    this.backend.createCollection(this.user.user, this.form.value)
      .subscribe(collection => this.router.navigate(['..', collection.name, 'settings'], {relativeTo: this.route}));
  }
}
