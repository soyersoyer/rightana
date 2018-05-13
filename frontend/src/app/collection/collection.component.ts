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

  getCollections(): Observable<Collection[]> {
    return this.backend.getCollections()
      .do(collections => this.collections = collections);
  }


  navigateToCreateIfEmpty() {
    if (this.collections && this.collections.length === 0) {
     this.router.navigate(['create'], {relativeTo: this.route});
    }
  }
}
