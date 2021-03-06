import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { Collection, Shard, BackendService } from '../backend.service';
import { UserComponent } from '../user/user.component';
import { RValidators } from '../forms/rvalidators';
import { ToastyService } from '../toasty/toasty.service';

@Component({
  selector: 'rana-collection-settings',
  templateUrl: './settings.component.html',
})
export class CollectionSettingsComponent implements OnInit {
  form: FormGroup;
  collection: Collection;

  shards: Shard[];
  allSize: number;

  constructor(
    private backend: BackendService,
    private router: Router,
    private route: ActivatedRoute,
    private fb: FormBuilder,
    private user: UserComponent,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      id: [null],
      name: [null, RValidators.collectionName],
    });
    this.route.parent.params.forEach((params: Params) => {
      const collectionName = params['collectionName'];
      this.getCollection(collectionName);
      this.getCollectionShards(collectionName);
    });
  }

  getCollection(collectionName: string) {
    this.backend
      .getCollection(this.user.user, collectionName)
      .subscribe(collection => {
        this.form.setValue(collection);
        this.collection = collection;
      });
  }

  getCollectionShards(collectionName: string) {
   this.backend
      .getCollectionShards(this.user.user, collectionName)
      .subscribe(shards => {
        this.shards = shards;
        this.allSize = shards.reduce((a, b) => a + b.size, 0);
      });
  }

  save() {
    this.backend.saveCollection(this.user.user, this.collection.name, this.form.value)
      .subscribe(_ => {
        this.toasty.success('Save success');
        this.router.navigate(['../..', this.form.value.name, 'settings'], {relativeTo: this.route});
      });
  }

  delete() {
    this.backend.deleteCollection(this.user.user, this.collection.name)
      .subscribe(_ => {
        this.toasty.success('Delete success');
        this.router.navigate(['../..'], {relativeTo: this.route});
      });
  }

  deleteShard(shard: Shard) {
    this.backend
      .deleteCollectionShard(this.user.user, this.collection.name, shard.id)
      .subscribe(_ => {
        this.getCollectionShards(this.collection.name);
      })
  }

}
