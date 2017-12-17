import {PipeTransform, Pipe} from '@angular/core';

@Pipe({name: 'bytes'})
export class BytesPipe implements PipeTransform {

  transform(value: number): string | number {
    const dictionary: Array<{max: number, trunc?: number, type: string}> = [
      { max: 1e3, type: 'B' },
      { max: 1e6, trunc: 1e1, type: 'KB' },
      { max: 1e9, trunc: 1e4, type: 'MB' },
      { max: 1e12, trunc: 1e7, type: 'GB' }
    ];

    const format = dictionary.find(d => value < d.max) || dictionary[dictionary.length - 1];
    const num = (format.trunc? value - (value % format.trunc) : value) / (format.max / 1e3);
    return `${num} ${format.type}`;
  }
}
