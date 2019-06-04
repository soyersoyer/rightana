import { Component, Input } from '@angular/core';

import { CollectionStatComponent } from './stat.component';

@Component({
  selector: 'rana-table-sum',
  templateUrl: './table-sum.component.html'
})
export class TableSumComponent {
  @Input() name: string;
  @Input() sums: any;
  @Input() key: string;

  showAll = false;

  constructor(
    private statComponent: CollectionStatComponent,
  ) {}

  showMeAll() {
    this.showAll = true;
  }

  addFilter(value: string) {
    this.statComponent.dashboardSetup.add(this.key, value);
  }

  removeFilter() {
    this.statComponent.dashboardSetup.del(this.key);
  }

  get inFilter(): boolean {
    return this.key && this.statComponent.setup.in(this.key);
  }
}
