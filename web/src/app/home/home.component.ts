import { Component, OnInit } from '@angular/core';

import { BackendService } from '../backend.service';

@Component({
  selector: 'rana-home',
  templateUrl: './home.component.html',
})
export class HomeComponent implements OnInit {

  constructor(
    private backend: BackendService,
  ) { }

  ngOnInit() {
  }

  get serverAnnounce(): string {
    return this.backend.config && this.backend.config.server_announce;
  }

}
