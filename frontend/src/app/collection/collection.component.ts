import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { Observable } from 'rxjs/Observable';
import { tap } from 'rxjs/operators';

import { BackendService, CollectionSummary } from '../backend.service';

@Component({
  selector: 'rana-collection',
  templateUrl: './collection.component.html',
})
export class CollectionComponent implements OnInit {
  collections: CollectionSummary[];

  constructor(
    private backend: BackendService,
    private route: ActivatedRoute,
    private router: Router,
  ) { }

  ngOnInit() {
    this.getCollections()
      .subscribe(_ => {
        if (this.route.snapshot.children.length === 0) {
          this.navigateToCreateIfEmpty();
        };
      });
  }

  getCollections(): Observable<CollectionSummary[]> {
    return this.backend.getCollectionSummaries()
      .do(collections => this.collections = collections);
  }


  navigateToCreateIfEmpty() {
    if (this.collections && this.collections.length === 0) {
     this.router.navigate(['create'], {relativeTo: this.route});
    }
  }
}
