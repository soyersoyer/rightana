import { Component, OnInit } from '@angular/core';

import { Backup, BackendService } from '../backend.service';
import { ToastyService } from '../toasty/toasty.module';

@Component({
  selector: 'rana-admin-backups',
  templateUrl: './admin-backups.component.html',
})
export class AdminBackupsComponent implements OnInit {
  backups: Backup[];

  constructor(
    private backend: BackendService,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.getUsers();
  }

  getUsers() {
    this.backend.getBackups().subscribe(backups => this.backups = backups)
  }

  run(b: Backup) {
    this.backend.runBackup(b.id)
      .subscribe(_ => this.toasty.success(`Backup (${b.id}) success`));
  }

  get origin(): string {
    return window.location.origin;
  }
}
