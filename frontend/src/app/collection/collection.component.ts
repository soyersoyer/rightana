import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { Observable } from 'rxjs/Observable';
import { tap } from 'rxjs/operators';

import { BackendService, Collection } from '../backend.service';

@Component({
  selector: 'k20a-collection',
  templateUrl: './collection.component.html',
})
export class CollectionComponent implements OnInit {
  collections: Collection[];

  constructor(
    private backend: BackendService,
    private router: Router,
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.getCollections()
      .subscribe(_ => {
        if (this.route.snapshot.children.length === 0) {
          this.navigateToFirst();
        }
      });
  }

  navigateToFirst() {
    if (this.collections && this.collections.length !== 0) {
      this.router.navigate([this.collections[0].id], {relativeTo: this.route});
    } else {
      this.router.navigate(['create'], {relativeTo: this.route});
    }
  }

  getCollections(): Observable<Collection[]> {
    return this.backend.getCollections()
      .do(collections => this.collections = collections);
  }

  create(formData: any) {
    this.backend.createCollection(formData)
      .subscribe(collection => {
        this.getCollections()
          .subscribe(_ => this.router.navigate([collection.id, 'settings'], {relativeTo: this.route}));
      });
  }

  save(formData: any) {
    this.backend.saveCollection(formData)
      .subscribe(collection => {
        this.getCollections()
          .subscribe(_ => this.router.navigate([collection.id], {relativeTo: this.route}));
      });
  }

  delete(collectionId: string) {
   this.backend.deleteCollection(collectionId)
      .subscribe(_ => {
        this.getCollections()
          .subscribe(__ => this.navigateToFirst());
      });
  }

}
