import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';

import { CollectionComponent } from '../collection/collection.component';
import { BackendService } from '../backend.service';

@Component({
  selector: 'k20a-collection-create',
  templateUrl: './create.component.html',
})
export class CollectionCreateComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private backend: BackendService,
    private router: Router,
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, [Validators.required]],
    });
  }

  create() {
    this.backend.createCollection(this.form.value)
      .subscribe(collection => this.router.navigate(['..', collection.id, 'settings'], {relativeTo: this.route}));
  }
}
