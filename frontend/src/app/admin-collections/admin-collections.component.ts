import { Component, OnInit } from '@angular/core';

import { CollectionInfo, BackendService } from '../backend.service';

@Component({
  selector: 'rana-admin-collections',
  templateUrl: './admin-collections.component.html',
})
export class AdminCollectionsComponent implements OnInit {
  collections: CollectionInfo[];

  constructor(
    private backend: BackendService,
  ) { }

  ngOnInit() {
    this.getUsers();
  }

  getUsers() {
    this.backend.getCollections().subscribe(collections => this.collections = collections)
  }
}
