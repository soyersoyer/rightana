import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { CollectionComponent } from '../collection/collection.component';

@Component({
  selector: 'k20a-collection-create',
  templateUrl: './create.component.html',
})
export class CollectionCreateComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private parent: CollectionComponent,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, [Validators.required]],
    });
  }

  create() {
    this.parent.create(this.form.value);
  }

}
