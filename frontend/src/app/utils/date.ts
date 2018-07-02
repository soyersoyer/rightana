export function getDateStrFromUnixTime(unix: number, res: string): string {
    var d = new Date(unix*1000);
    var datestr = d.getFullYear()+"-"+padNumber(d.getMonth()+1)+"-"+padNumber(d.getDate());
    if (res === "hour") {
      return datestr+" "+padNumber(d.getHours())+":"+padNumber(d.getMinutes());
    }
    return datestr;
}

function padNumber(n: number): string {
    return n<10?"0"+n:""+n
}
