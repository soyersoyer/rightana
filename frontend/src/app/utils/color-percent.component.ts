import { Component, Input } from '@angular/core';

@Component({
  selector: 'k20a-color-percent',
  templateUrl: './color-percent.component.html'
})
export class ColorPercentComponent {
  @Input() percent: number;
  @Input() inverse = false;

  getClass() {
    const p = this.inverse ? this.percent * -1 : this.percent;
    if (p < 0) {
      return 'text-danger';
    }
    if (p > 0) {
      return 'text-success';
    }
    return '';
  }
}
