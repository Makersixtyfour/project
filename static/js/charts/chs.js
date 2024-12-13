const drawAssemblyVTChart = (data, plandata, inventorydata, target) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 80};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  let flag = true;
  data.sort((a, b) => a.type.localeCompare(b.type))
  const series = d3.stack().keys(d3.union(data.map(d => d.type))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).value)(d3.index(data, d => d.date, d => d.type))
  const planseries = d3.stack().keys(d3.union(plandata.map(d => d.plantype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).plan)(d3.index(plandata, d => d.date, d => d.plantype))
  const x = d3.scaleBand().domain(plandata.map(d => d.date)).range([0, innerWidth]).padding(0.1);
  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()
  const color = d3.scaleOrdinal().domain(["X1-brand", "X1-rh", "X2-brand", "X2-rh", "X1-white", "X2-white", "brand", "rh", "X7-brand", "X7-rh", "X3-brand", "X3-rh"]).range(["#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE", "#D1D1D1", "#D1D1D1", "#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE"]).unknown("white");
  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])
  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)
  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0]) + x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - 5 : y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", 2*x.bandwidth()/3).on("mouseover", e => {flag = !flag;change(flag);}).append("title").text(d => d[1] - d[0])
  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))
  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 11).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[1]) - 15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).text(d => `${d3.format(",.0f")(d[1])}` )
  series.forEach(serie => {innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("class", d => d.key.substring(0,2)).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - (y(d[1]) - y(d[0]))/2 - 5 : y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", d => d.key.startsWith("X1") ? "#921A40" : "#102C57").text(d => {if (d[1] - d[0] >= 3500) { return `${d3.format(",.0f")(d[1]-d[0])}` }})})

  const x1data = series.filter(s => s.key.startsWith("X1"))
  const x1rhdata = x1data[x1data.length-1]
  if (x1rhdata != undefined) {innerChart.append("g").selectAll().data(x1rhdata).join("text").text(d => d[1] > 3500 ? `Σ${d3.format(",.0f")(d[1])}` : "").attr("class", "x1total").attr("text-anchor", "middle").attr("dominant-baseline", "hanging").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[0])).attr("dy", "0.1em").attr("fill", "#921A40").attr("font-size", 12).attr("opacity", 0)
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    }
  }
  svg.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#A5A0DE").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 20).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("White").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 35).attr("dy", "0.35em").attr("fill", "#D1D1D1").attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("Xưởng 3").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 50).attr("dy", "0.35em").attr("fill", "#102C57").attr("font-weight", 600).attr("font-size", 12) 

  svg.append("text").text("Xưởng 7").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 65).attr("dy", "0.35em").attr("fill", "#921A40").attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 80).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12)

  svg.append("text").text(`Total: `).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 80).attr("y", 8).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 14).append("tspan").text(`$${d3.format(",.0f")(d3.sum(data, d => d.value))}`).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("fill", "#75485E").attr("font-size", 16).attr("font-weight", 600)

  svg.append("text").text(" <-- Giá trị total actual dựa theo số ngày trên chart").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 200).attr("y", 8).attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  svg.append("text").text("* Rà chuột vào cột để hiện value cho loại hàng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 80).attr("y", 25).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  innerChart.selectAll().data(planseries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/3).attr("stroke", "#FF9874").on("mouseover", e => {flag = !flag;change(flag);})

  const diffs = d3.rollups(plandata, D => { return {"current": D[0].plan + D[1].plan, "prev": D[0].plan + D[1].plan + D[0].change + D[1].change}} ,d => d.date)
  innerChart.selectAll().data(diffs).join("rect").attr("x", d => x(d[0])).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) : y(d[1].prev)).attr("height", d => y(0) - y(Math.abs(d[1].current-d[1].prev))).attr("width", x.bandwidth()/3).attr("fill", "url(#diffpattern)").attr("fill-opacity", 0.6).append("title").text(d => Math.abs(d[1].current-d[1].prev))

  innerChart.selectAll().data(diffs).join("text").text(d => {if (d[1].current > d[1].prev) {return "︽"}if (d[1].current < d[1].prev) { return "︾"}}).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d[0]) + x.bandwidth()/6).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) + 10 : y(d[1].prev) + 10).attr("font-weight", 900).attr("fill", d => {if (d[1].current > d[1].prev) {return "#3572EF"}if (d[1].current < d[1].prev) {return "#C80036"}})
  
  planseries.forEach(planserie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(planserie).join("text").text(d => `${d3.format(",.0f")(d[1]-d[0])}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/6).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", "#102C57").attr("transform", d => `rotate(-90, ${x(d.data[0]) + x.bandwidth()/6}, ${y(d[1]) - (y(d[1]) - y(d[0]))/2})`)
  })

  innerChart.append("text").text("plan->").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", x(plandata[plandata.length-1].date) + x.bandwidth()/6).attr("y", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 20).attr("fill", "#FA7070").attr("font-size", 14).attr("transform", `rotate(90, ${x(plandata[plandata.length-1].date) + x.bandwidth()/6}, ${y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 20})`)

  if (target != undefined) { 
    const dates = Array.from(d3.union(plandata.map(d => d.date), data.map(d => d.date))) 
    target = target.filter(t => dates.includes(t.date))
    innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)
    innerChart.append("g").selectAll().data(target).join("text").attr("text-anchor", "end").attr("alignment-baseline", "middle").text(d => `- ${d3.format("~s")(d.value)}`).attr("font-size", "12px").attr("x", innerWidth + 10).attr("y", d => y(d.value)).attr("fill", "currentColor")

    innerChart.append("text").text("Target").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", y(d3.min(target, d => d.value)) + 10).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(-90, ${innerWidth}, ${y(d3.min(target, d => d.value)) + 10})`)
  }

  if (inventorydata != undefined) {
    const y2 = d3.scaleLinear().domain([0, d3.sum(inventorydata, d => d.inventory)]).rangeRound([3*innerHeight/4, 0]).nice()
    const leftInnerChart = svg.append("g").attr("transform", `translate(0, ${innerHeight/4})`)
    svg.append("line").attr("x1", 70).attr("y1", height).attr("x2", 70).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)
    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[0].inventory) + margin.bottom).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[0].inventory)).attr("fill", color(inventorydata[0].type));
    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[1].inventory)).attr("fill", color(inventorydata[1].type));

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[2].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory)) - 5).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[2].inventory)).attr("fill", color(inventorydata[2].type));

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[3].inventory)).attr("fill", color(inventorydata[3].type));

    leftInnerChart.append("text").text((inventorydata[0].inventory != 0) ? `${d3.format(",.0f")(inventorydata[0].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[0].inventory/2) + margin.bottom).attr("fill", "#921A40").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[1].inventory != 0) ? `${d3.format(",.0f")(inventorydata[1].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[1].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#921A40").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[2].inventory != 0) ? `${d3.format(",.0f")(inventorydata[2].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[2].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory)) - 5).attr("fill", "#102C57").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[3].inventory != 0) ? `${d3.format(",.0f")(inventorydata[3].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[3].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("fill", "#102C57").attr("font-size", 12)

    leftInnerChart.append("text").text(`${d3.format(",.0f")(d3.sum(inventorydata, d => d.inventory))}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)

    leftInnerChart.append("text").text(inventorydata[0].createdat).attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 60).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(90, 60, ${y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5})`)

    svg.append("text").text("Inventory").attr("text-anchor", "start").attr("x", 5).attr("y", height).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)
  } 

  return svg.node();

  function change(flag) {
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    }
  }
}


const drawAssemblyEfficiencyChart = (data, manhr) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  const dates = new Set(data.map(d => d.date)) 
  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.value)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.value)).attr("height", d => y(0) - y(d.value)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2")

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d3.format(",.0f")(d.value)).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.value)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.value)})`)
      
  svg.append("text").text("Sản lượng($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).value / w.workhr / 57.5 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g")
    .selectAll()
    .data(workinghrs)
    .join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  if (workinghrs.length != 0) {  
    const lastW = workinghrs[workinghrs.length-1]
    innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)
  }
  svg.append("text").text("Demand: 57.5 $/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)}

  return svg.node();
}

const drawCuttingChart2 = (data, returndata, finedata, target_actual, prodtypedata, target) => {
  if (returndata != undefined) {
    data = data.map(d => {
      d.return = 0;
      returndata.forEach(rd => {
        if (rd.date == d.date && rd.is25 == d.is25) {
          d.qty = d.qty - rd.qty;
          d.return = rd.qty;
        }
      });
      return d;
    })
  }

  if (target == undefined) {
    target = [{"date": "", "value": 0}]
  }
  const width = 900;
  const height = 350;
  const margin = {top: 30, right: 20, bottom: 20, left: 150};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.is25))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.is25))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const maxReturn = (returndata != undefined) ? d3.max(d3.rollup(returndata, D => d3.sum(D, d => d.qty) ,d => d.date), d => d[1]) + 2 : 0;

  const y = d3.scaleLinear().domain([-maxReturn,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(["#E4E0E1", "#FFBB70", "#A0D9DE"]).unknown("#ccc");

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 1).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/2)

  if (finedata != undefined) {
    const fineseries = d3.stack().keys(d3.union(finedata.map(d => d.is25reeded))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(finedata, d => d.date, d => d.is25reeded))

    const color1 = d3.scaleOrdinal().domain(fineseries.map(d => d.key)).range(["#E4E0E1", "#FFBB70", "#A0D9DE"]).unknown("#ccc");

    innerChart.selectAll().data(fineseries).join("g").attr("fill", d => color1(d.key)).attr("fill-opacity", 1).attr("stroke", "#00CCDD").selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])+ x.bandwidth()/2).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/2)

    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(fineseries[fineseries.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 3*x.bandwidth()/4).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).text(d => `${d3.format(".1f")(d[1])}`)

    fineseries.forEach(serie => {
      innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 3*x.bandwidth()/4).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("fill", "#75485E").attr("font-size", "12px").text(d => {if (d[1] - d[0] >= 1.5) { return d3.format(".1f")(d[1]-d[0])}})
    })}

  if (returndata != undefined) {
    const returnseries = d3.stack().keys(d3.union(returndata.map(d => d.is25))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(returndata, d => d.date, d => d.is25))
    
    const y1 = d3.scaleLinear().domain([0, maxReturn]).rangeRound([y(0), innerHeight]).nice() 
      
    innerChart.selectAll().data(returnseries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 1).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y1(d[0])).attr("height", d => y1(d[1]-d[0]) - y(0)).attr("width", x.bandwidth()/2).append("title").text(d => d[1] - d[0])

    returnseries.forEach(serie => {
      innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/4).attr("y", d => y1(d[0] + (d[1]-d[0])/2)).attr("fill", "#75485E").attr("dy", "0.15em").attr("font-size", "12px").text(d => {
        if (d[1] - d[0] >= 1) { return d3.format(".1f")(d[1]-d[0])}})
    })

    innerChart.append("line").attr("x1", 0).attr("y1", y(0)).attr("x2", innerWidth).attr("y2", y(0)).attr("stroke", "black")
    innerChart.append("line").attr("x1", innerWidth).attr("y1", y(0)).attr("x2", innerWidth).attr("y2", innerHeight).attr("stroke", "black")
    svg.append("text").text("Nhập lại kho").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", width-10).attr("y", innerHeight - 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", `rotate(90, ${width-10}, ${innerHeight - 5})`)
  }
  
  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/4).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).text(d => `${d3.format(".1f")(d[1])}`)

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/4).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").text(d => {if (d[1] - d[0] >= 1.5) { return d3.format(".1f")(d[1]-d[0])}})
  })

 innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {if (i == target.length-1) return d.value;else {if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 2) return d.value;}}).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)

svg.append("text").text("Sản lượng (m³): ").attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 140).attr("y", height).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 12).attr("transform", `rotate(-90, 140, ${height})`).append("tspan").text("Gỗ 25mm Reeded").attr("fill", color(true)).attr("font-weight", 600).append("tspan").text(", Gỗ Còn Lại").attr("fill", color(false)).attr("font-weight", 600).append("tspan").text(", Gỗ tinh").attr("fill", "#00CCDD").attr("font-weight", 600).append("tspan").text(", ").attr("font-weight", 300).attr("fill", "#75485E").append("tspan").text("Target").attr("fill", "#FA7070").attr("font-weight", 600)

const y1 = d3.scaleLinear().domain([0,  d3.max(target_actual.detail.map(d => d.target))]).rangeRound([innerHeight, 0]).nice()

svg.append("text").text(`${target_actual.name}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 65).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14)

svg.append("text").text(`Tổng (m³) Gỗ Tinh từ ${target_actual.startdate} --> ${target_actual.enddate}`).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 7).attr("y", height).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14).attr("transform", `rotate(-90, 7, ${height})`)

svg.append("line").attr("x1", 120).attr("y1", height).attr("x2", 120).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)

svg.append("rect").attr("x", 20).attr("y", y1(target_actual.detail[0].target) + margin.top).attr("width", 40).attr("height", innerHeight - y1(target_actual.detail[0].target)).attr("stroke", "black").attr("stroke-width", "1px").attr("fill", "transparent")

svg.append("text").text(`${target_actual.detail[1].type}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 40).attr("y", height - 10).attr("fill", "#102C57").attr("font-size", 10)

svg.append("text").text(`${target_actual.detail[0].target}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 40).attr("y", y1(target_actual.detail[0].target) + margin.top - 5).attr("fill", "#102C57").attr("font-size", 12)

svg.append("rect").attr("x", 70).attr("y", y1(target_actual.detail[1].target) + margin.top).attr("width", 40).attr("height", innerHeight - y1(target_actual.detail[1].target)).attr("stroke", "black").attr("stroke-width", "1px").attr("fill", "transparent")

svg.append("text").text(`${target_actual.detail[0].type}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 90).attr("y", height - 10).attr("fill", "#102C57").attr("font-size", 10)

svg.append("text").text(`${target_actual.detail[1].target}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 90).attr("y", y1(target_actual.detail[1].target) + margin.top - 5).attr("fill", "#102C57").attr("font-size", 12)

if (prodtypedata == undefined) {
  prodtypedata = [
    {"type": false, qty: 0}, {"type": true, qty: 0}
  ]
}

svg.append("rect").attr("x", 20 + 2.5).attr("y", y1(prodtypedata[0].qty) + margin.top).attr("width", 40 - 5).attr("height", innerHeight - y1(prodtypedata[0].qty)).attr("fill", "#E4E0E1")

svg.append("rect").attr("x", 70 + 2.5).attr("y", y1(prodtypedata[1].qty) + margin.top).attr("width", 40 - 5).attr("height", innerHeight - y1(prodtypedata[1].qty)).attr("fill", "#FFBB70")

svg.append("text").text(prodtypedata[0].qty > 0 ? `${d3.format(".0f")(prodtypedata[0].qty)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 40).attr("y", y1(prodtypedata[0].qty) + margin.top - 5).attr("fill", "#102C57").attr("font-size", 12)

svg.append("text").text(prodtypedata[1].qty > 0 ? `${d3.format(".0f")(prodtypedata[1].qty)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 90).attr("y", y1(prodtypedata[1].qty) + margin.top - 5).attr("fill", "#102C57").attr("font-size", 12)
    
svg.append("text").text(prodtypedata[0].qty > 0 ? `${d3.format(".0f")(prodtypedata[0].qty / target_actual.detail[0].target * 100)}%` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 40).attr("y", y1(prodtypedata[0].qty/2) + margin.top).attr("fill", "#102C57").attr("font-size", 12)
svg.append("text").text(prodtypedata[1].qty > 0 ? `${d3.format(".0f")(prodtypedata[1].qty / target_actual.detail[1].target * 100)}%` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 90).attr("y", y1(prodtypedata[1].qty/2) + margin.top).attr("fill", "#102C57").attr("font-size", 12)

  return svg.node();
}


const drawCuttingChart3 = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 10, right: 10, bottom: 20, left: 30};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const series = d3.stack().keys(d3.union(data.map(d => d.prodtype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.prodtype))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/2]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(["#A5A0DE", "#DFC6A2", "#A0D9DE"]).unknown("#ccc");

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 1).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/2)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/4).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).text(d => `Σ${d3.format(".1f")(d[1])}`)

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/4).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").text(d => {  if (d[1] - d[0] >= 1) { return d3.format(".1f")(d[1]-d[0])}})
  })
 
if (manhr != undefined) {
  const workinghrs = manhr.filter(d => dates.has(d.date))
  
  const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

  innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
    
  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4}, ${y1(d.workhr)})`)

const tmp = series[series.length-1]
workinghrs.forEach(w => {
  w.efficiency = tmp.filter(d => d.data[0] == w.date)[0][1]  / w.workhr / 0.03 * 100
})

const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "12px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

const lastW = workinghrs[workinghrs.length-1]
innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 15).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("với ").attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 5).attr("y", 7).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).append("tspan").text("Demand: 0.03 m³/h").attr("fill", "#75485E").attr("font-weight", 600)}

svg.append("text").text("Sản lượng (m³) cắt cho hàng ").attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 10).attr("y", height-margin.bottom).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).attr("transform", `rotate(-90, 10, ${height-margin.bottom})`).append("tspan").text("Brand").attr("fill", color("brand")).attr("font-weight", 600).append("tspan").text(", RH").attr("fill", color("rh")).attr("font-weight", 600).append("tspan").text(" và ").attr("font-weight", 300).attr("fill", "#75485E").append("tspan").text(" Nhân lực").attr("fill", "#90D26D").attr("font-weight", 600)

  return svg.node();
}

const drawLaminationChart1 = (data, target) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.prodtype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.prodtype))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(["#DFC6A2", "#A5A0DE", "#A0D9DE"]).unknown("#ccc");

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 1).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()).append("title").text(d => d[1]-d[0])

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("font-weight", 600).text(d => `Σ ${d3.format(".3s")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 - 4).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").text(d => {if (d[1] - d[0] >= 30) { return d3.format(".3s")(d[1]-d[0])}})
  })

if (target != undefined) { 
  const dates = data.map(d => d.date)
target = target.filter(t => dates.includes(t.date))
innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {
  if (i == target.length-1) return d.value;
  else {
    if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 15) return d.value;
  }
  }).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)
}

const maxOne = series[1].find(d => d[1] == d3.max(series[1], d => d[1]))

svg.append("text").text("Sản lượng (m²) cho hàng ").attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 10).attr("y", height-margin.bottom).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).attr("transform", `rotate(-90, 10, ${height-margin.bottom})`).append("tspan").text("BRAND").attr("fill", color("brand")).attr("font-weight", 600).append("tspan").text(", RH").attr("fill", color("rh")).attr("font-weight", 600).append("tspan").text(" với ").attr("font-weight", 300).attr("fill", "#75485E").append("tspan").text(" Target").attr("fill", "#FA7070").attr("font-weight", 600)

  return svg.node();
}

const drawLaminationChart2 = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("height", d => y(0) - y(d.qty)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2")

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d3.format(".3s")(d.qty)).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.qty)})`)
      
svg.append("text").text("Sản lượng(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).qty / w.workhr / 2 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".3s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  const lastW = workinghrs[workinghrs.length-1]
  innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Demand: 2 m²/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

function Legend(color, {
  title,
  tickSize = 6,
  width = 320, 
  height = 44 + tickSize,
  marginTop = 18,
  marginRight = 0,
  marginBottom = 16 + tickSize,
  marginLeft = 0,
  ticks = width / 64,
  tickFormat,
  tickValues
  } = {}) {

  function ramp(color, n = 256) {
    const canvas = document.createElement("canvas");
    canvas.width = n;
    canvas.height = 1;
    const context = canvas.getContext("2d");
    for (let i = 0; i < n; ++i) {
      context.fillStyle = color(i / (n - 1));
      context.fillRect(i, 0, 1, 1);
    }
    return canvas;
  }

  const svg = d3.create("svg")
      .attr("width", width)
      .attr("height", height)
      .attr("viewBox", [0, 0, width, height])
      .style("overflow", "visible")
      .style("display", "block");

  let tickAdjust = g => g.selectAll(".tick line").attr("y1", marginTop + marginBottom - height);
  let x;

  // Continuous
  if (color.interpolate) {
    const n = Math.min(color.domain().length, color.range().length);

    x = color.copy().rangeRound(d3.quantize(d3.interpolate(marginLeft, width - marginRight), n));

    svg.append("image")
        .attr("x", marginLeft)
        .attr("y", marginTop)
        .attr("width", width - marginLeft - marginRight)
        .attr("height", height - marginTop - marginBottom)
        .attr("preserveAspectRatio", "none")
        .attr("xlink:href", ramp(color.copy().domain(d3.quantize(d3.interpolate(0, 1), n))).toDataURL());
  }

  // Sequential
  else if (color.interpolator) {
    x = Object.assign(color.copy()
        .interpolator(d3.interpolateRound(marginLeft, width - marginRight)),
        {range() { return [marginLeft, width - marginRight]; }});

    svg.append("image")
        .attr("x", marginLeft)
        .attr("y", marginTop)
        .attr("width", width - marginLeft - marginRight)
        .attr("height", height - marginTop - marginBottom)
        .attr("preserveAspectRatio", "none")
        .attr("xlink:href", ramp(color.interpolator()).toDataURL());

    // scaleSequentialQuantile doesn’t implement ticks or tickFormat.
    if (!x.ticks) {
      if (tickValues === undefined) {
        const n = Math.round(ticks + 1);
        tickValues = d3.range(n).map(i => d3.quantile(color.domain(), i / (n - 1)));
      }
      if (typeof tickFormat !== "function") {
        tickFormat = d3.format(tickFormat === undefined ? ",f" : tickFormat);
      }
    }
  }

  // Threshold
  else if (color.invertExtent) {
    const thresholds
        = color.thresholds ? color.thresholds() // scaleQuantize
        : color.quantiles ? color.quantiles() // scaleQuantile
        : color.domain(); // scaleThreshold

    const thresholdFormat
        = tickFormat === undefined ? d => d
        : typeof tickFormat === "string" ? d3.format(tickFormat)
        : tickFormat;

    x = d3.scaleLinear()
        .domain([-1, color.range().length - 1])
        .rangeRound([marginLeft, width - marginRight]);

    svg.append("g")
      .selectAll("rect")
      .data(color.range())
      .join("rect")
        .attr("x", (d, i) => x(i - 1))
        .attr("y", marginTop)
        .attr("width", (d, i) => x(i) - x(i - 1))
        .attr("height", height - marginTop - marginBottom)
        .attr("fill", d => d);

    tickValues = d3.range(thresholds.length);
    tickFormat = i => thresholdFormat(thresholds[i], i);
  }

  // Ordinal
  else {
    x = d3.scaleBand()
        .domain(color.domain())
        .rangeRound([marginLeft, width - marginRight]);

    svg.append("g")
      .selectAll("rect")
      .data(color.domain())
      .join("rect")
        .attr("x", x)
        .attr("y", marginTop)
        .attr("width", Math.max(0, x.bandwidth() - 1))
        .attr("height", height - marginTop - marginBottom)
        .attr("fill", color);

    tickAdjust = () => {};
  }

  svg.append("g")
      .attr("transform", `translate(0,${height - marginBottom})`)
      .call(d3.axisBottom(x)
        .ticks(ticks, typeof tickFormat === "string" ? tickFormat : undefined)
        .tickFormat(typeof tickFormat === "function" ? tickFormat : undefined)
        .tickSize(tickSize)
        .tickValues(tickValues))
      .call(tickAdjust)
      .call(g => g.select(".domain").remove())
      .call(g => g.append("text")
        .attr("x", marginLeft)
        .attr("y", marginTop + marginBottom - height - 6)
        .attr("fill", "currentColor")
        .attr("text-anchor", "start")
        .attr("font-weight", "bold")
        .attr("class", "title")
        .text(title));

  return svg.node();     
}

function Swatches(color, {
  columns = null,
  format,
  unknown: formatUnknown,
  swatchSize = 15,
  swatchWidth = swatchSize,
  swatchHeight = swatchSize,
  marginLeft = 0
} = {}) {
  const id = `-swatches-${Math.random().toString(16).slice(2)}`;
  const unknown = formatUnknown == null ? undefined : color.unknown();
  const unknowns = unknown == null || unknown === d3.scaleImplicit ? [] : [unknown];
  const domain = color.domain().concat(unknowns);
  if (format === undefined) format = x => x === unknown ? formatUnknown : x;

  function entity(character) {
    return `&#${character.charCodeAt(0).toString()};`;
  }

  if (columns !== null) return htl.html`<div style="display: flex; align-items: center; margin-left: ${+marginLeft}px; min-height: 33px; font: 10px sans-serif;">
  <style>

.${id}-item {
  break-inside: avoid;
  display: flex;
  align-items: center;
  padding-bottom: 1px;
}

.${id}-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: calc(100% - ${+swatchWidth}px - 0.5em);
}

.${id}-swatch {
  width: ${+swatchWidth}px;
  height: ${+swatchHeight}px;
  margin: 0 0.5em 0 0;
}

  </style>
  <div style=${{width: "100%", columns}}>${domain.map(value => {
    const label = `${format(value)}`;
    return htl.html`<div class=${id}-item>
      <div class=${id}-swatch style=${{background: color(value)}}></div>
      <div class=${id}-label title=${label}>${label}</div>
    </div>`;
  })}
  </div>
</div>`;

  return htl.html`<div style="display: flex; align-items: center; min-height: 33px; margin-left: ${+marginLeft}px; font: 10px sans-serif;">
  <style>

.${id} {
  display: inline-flex;
  align-items: center;
  margin-right: 1em;
}

.${id}::before {
  content: "";
  width: ${+swatchWidth}px;
  height: ${+swatchHeight}px;
  margin-right: 0.5em;
  background: var(--color);
}

  </style>
  <div>${domain.map(value => htl.html`<span class="${id}" style="--color: ${color(value)}">${format(value)}</span>`)}</div>`;
}

const drawRdOpTotalChart = (data, inventory) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 120, bottom: 20, left: 40};
  let innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  data = data.filter(d => d.section !== "Inventory")

  if (inventory == undefined) {
    margin.right = 20
    innerWidth = width - margin.left - margin.right;
  }

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height]);
  
  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`);

  const x = d3.scaleBand().domain(data.map(d => d.section)).range([0, innerWidth]).paddingInner(0.2);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).range([innerHeight, innerHeight/4]).nice();

  const y1 = d3.scaleLinear().domain(d3.extent(data, d => d.avg)).range([innerHeight/4, 0]).nice();

  const y2 = d3.scaleLinear().domain([0, inventory?.qty]).range([innerHeight, 0]).nice();

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px").attr("font-weight", 600).style("text-transform", "capitalize"))

  innerChart.selectAll(`rect`).data(data).join("rect").attr("x", d => x(d.section)).attr("y", d => y(d.qty)).attr("width", x.bandwidth()).attr("height", d => y(0) - y(d.qty)).attr("fill", "#DCA47C");

  innerChart.append("g").attr("font-family", "san-serif").attr("font-size", 14).attr("font-weight", 600).selectAll().data(data).join("text").text(d => d3.format(",.3s")(d.qty)).attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("x", d => x(d.section) + x.bandwidth()/2).attr("y", d => y(d.qty)).attr("dy", "-0.1em").attr("fill", "#75485E")

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.section) + x.bandwidth()/2).y(d => y1(d.avg)).curve(d3.curveStep)(data));

  innerChart.append("g").attr("font-family", "san-serif").attr("font-size", 14).attr("font-weight", 600).selectAll().data(data).join("text").text(d => d3.format(",.3s")(d.avg)).attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("x", d => x(d.section) + x.bandwidth()/2).attr("y", d => y1(d.avg)).attr("dy", "-0.1em").attr("fill", "#75485E")

  innerChart.append("text").text("AVG: ").attr("font-size", "14px").attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("fill", "#75485E").attr("x", x(data[0].section) + x.bandwidth()/2 - 30).attr("y", y1(data[0].avg)).attr("dy", "-0.1em").attr("font-weight", 600)

  svg.append("text").text(`Tổng sản lượng (m²) `).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 20).attr("y", height).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).attr("transform", `rotate(-90, 10, ${height})`).append("tspan").text(`${data[0].type.toUpperCase()}`).attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 16).append("tspan").text(` tới ngày `).attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).append("tspan").text(`${data[0].lastdate}`).attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 16)

  if (inventory != undefined) {
    svg.append("line").attr("x1", innerWidth + 3*x.bandwidth()/4).attr("y1", height).attr("x2", innerWidth + 3*x.bandwidth()/4).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)

    svg.append("rect").attr("x", innerWidth + x.bandwidth()).attr("y", y2(inventory.qty) + margin.bottom).attr("width", x.bandwidth()).attr("height", y2(0) - y2(inventory.qty)).attr("fill", "#DCA47C");

    svg.append("text").text(inventory.section).attr("text-anchor", "middle").attr("x", innerWidth + 3* x.bandwidth() /2).attr("y", height).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 14)

    svg.append("text").text(d3.format(",.0f")(inventory.qty)).attr("text-anchor", "middle").attr("x", innerWidth + 3* x.bandwidth() /2).attr("y", y2(inventory.qty) + margin.bottom).attr("dy", "-0.15em").attr("fill", "#75485E").attr("font-size", 14).attr("font-weight", 600)

    svg.append("text").text(inventory.qty != 0 ? inventory.date : "").attr("text-anchor", "middle").attr("alignment-baseline", "end").attr("x", innerWidth + 3*x.bandwidth()/4).attr("y", height/2).attr("dy", "1em").attr("fill", "#75485E").attr("font-size", 14).attr("transform", `rotate(-90, ${innerWidth + 3*x.bandwidth()/4}, ${height/2})`)
  } 

  return svg.node();
}


const drawValueTargetChart = (data, target) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  
  let flag = true;

  const series = d3.stack().keys(d3.union(data.map(d => d.type).sort())).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).value)(d3.index(data, d => d.date, d => d.type))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(["X1-brand", "X1-rh", "X2-brand", "X2-rh"]).range(["#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE"]).unknown("white");

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`).attr("transition", "all 1s")

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - 5 : y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()).on("mouseover", e => {flag = !flag;change(flag);}).append("title").text(d => d[1]-d[0])

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).text(d => `Σ ${d3.format(",.0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(serie).join("text").attr("class", d => d.key.substring(0,2)).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - (y(d[1]) - y(d[0]))/2 - 5 : y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", d => d.key.startsWith("X1") ? "#921A40" : "#102C57").text(d => {if (d[1] - d[0] >= 3500) { return `${d3.format(",.0f")(d[1]-d[0])}` }})
  })
  const x1data = series.filter(s => s.key.startsWith("X1"))
  const x1rhdata = x1data[x1data.length-1]
  if (x1rhdata != undefined) {
    innerChart.append("g").selectAll().data(x1rhdata).join("text").text(d => d[1] > 3500 ? `Σ ${d3.format(",.0f")(d[1])}`: "").attr("class", "x1total").attr("text-anchor", "middle").attr("dominant-baseline", "hanging").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[0])).attr("dy", "0.1em").attr("fill", "#921A40").attr("font-size", 14).attr("opacity", 0)

      if (flag) {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
      } else {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
      }

  }
  svg.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#A5A0DE").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 12)
     
  svg.append("text").text("Xưởng 2").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill", "#102C57").attr("font-weight", 600).attr("font-size", 12) 

  svg.append("text").text("Xưởng 1").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 80).attr("dy", "0.35em").attr("fill", "#921A40").attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 105).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12)

  svg.append("text").text("* Rà chuột vào cột để hiện value cho loại hàng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 40).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)
  
  if (target != undefined && target.length != 0) { 
    const dates = data.map(d => d.date)
    target = target.filter(t => dates.includes(t.date))
    innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

    innerChart.append("g").selectAll().data(target).join("text").attr("text-anchor", "end").attr("alignment-baseline", "middle").text(d => `- ${d3.format("~s")(d.value)}`).attr("font-size", "12px").attr("x", innerWidth + 20).attr("y", d => y(d.value))

    innerChart.append("text").text("Target").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", y(d3.min(target, d => d.value)) + 10).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(-90, ${innerWidth}, ${y(d3.min(target, d => d.value)) + 10})`)

  }

  return svg.node();

  function change(flag) {
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    }
  }
}

const drawPactVTChart = (data, plandata, inventorydata, target) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 80};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  
  let flag = true;

  data.sort((a, b) => a.type.localeCompare(b.type))

  const series = d3.stack().keys(d3.union(data.map(d => d.type))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).value)(d3.index(data, d => d.date, d => d.type))

  const planseries = d3.stack().keys(d3.union(plandata.map(d => d.plantype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).plan)(d3.index(plandata, d => d.date, d => d.plantype))

  const x = d3.scaleBand().domain(d3.union(data.map(d=> d.date), plandata.map(d => d.date))).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(["X1-brand", "X1-rh", "X2-brand", "X2-rh", "X1-stock", "X2-stock", "brand", "rh"]).range(["#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE", "#D1D1D1", "#D1D1D1", "#DFC6A2", "#A5A0DE"]).unknown("white");

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0]) + x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - 5 : y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", 2*x.bandwidth()/3).on("mouseover", e => {flag = !flag;change(flag);}).append("title").text(d => d[1] - d[0])

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 11).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[1]) - 15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).text(d => `${d3.format(",.0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("class", d => d.key.substring(0,2)).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - (y(d[1]) - y(d[0]))/2 - 5 : y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", d => d.key.startsWith("X1") ? "#921A40" : "#102C57").text(d => {if (d[1] - d[0] >= 3500) { return `${d3.format(",.0f")(d[1]-d[0])}` }})
  })

  const x1data = series.filter(s => s.key.startsWith("X1"))
  const x1rhdata = x1data[x1data.length-1]
  if (x1rhdata != undefined) {
    innerChart.append("g").selectAll().data(x1rhdata).join("text").text(d => d[1] > 3500 ? `Σ${d3.format(",.0f")(d[1])}` : "").attr("class", "x1total").attr("text-anchor", "middle").attr("dominant-baseline", "hanging").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[0])).attr("dy", "0.1em").attr("fill", "#921A40").attr("font-size", 12).attr("opacity", 0)

      if (flag) {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
      } else {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
      }
  }

  svg.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#A5A0DE").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 20).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 12)
     
  svg.append("text").text("Stock").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 35).attr("dy", "0.35em").attr("fill", "#D1D1D1").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Xưởng 2").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 50).attr("dy", "0.35em").attr("fill", "#102C57").attr("font-weight", 600).attr("font-size", 12) 

  svg.append("text").text("Xưởng 6").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 65).attr("dy", "0.35em").attr("fill", "#921A40").attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 80).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12)

  svg.append("text").text(`Total: `).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 95).attr("y", 8).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 14).append("tspan").text(`$${d3.format(",.0f")(d3.sum(data, d => d.value))}`).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("fill", "#75485E").attr("font-size", 16).attr("font-weight", 600)
  svg.append("text").text(" <-- Giá trị total actual dựa theo số ngày trên chart").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 200).attr("y", 8).attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  svg.append("text").text("* Rà chuột vào cột để hiện value cho loại hàng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 80).attr("y", 25).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  innerChart.selectAll().data(planseries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/3).attr("stroke", "#FF9874").on("mouseover", e => {flag = !flag;change(flag);}).append("title").text(d => d[1] - d[0])

  const diffs = d3.rollups(plandata, D => { return {"current": D[0].plan + D[1].plan, "prev": D[0].plan + D[1].plan + D[0].change + D[1].change}} ,d => d.date)
  innerChart.selectAll().data(diffs).join("rect").attr("x", d => x(d[0])).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) : y(d[1].prev)).attr("height", d => y(0) - y(Math.abs(d[1].current-d[1].prev))).attr("width", x.bandwidth()/3).attr("fill", "url(#diffpattern)").attr("fill-opacity", 0.3).append("title").text(d => Math.abs(d[1].current-d[1].prev))

  innerChart.selectAll().data(diffs).join("text").text(d => {if (d[1].current > d[1].prev) {return "︽"}if (d[1].current < d[1].prev) {return "︾"}}).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d[0]) + x.bandwidth()/6).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) + 10 : y(d[1].prev) + 10).attr("font-weight", 900).attr("fill", d => {if (d[1].current > d[1].prev) {return "#3572EF"}if (d[1].current < d[1].prev) {return "#C80036"}})

  planseries.forEach(planserie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(planserie).join("text").text(d => `${d3.format(",.0f")(d[1]-d[0])}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/6).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", "#102C57").attr("transform", d => `rotate(-90, ${x(d.data[0]) + x.bandwidth()/6}, ${y(d[1]) - (y(d[1]) - y(d[0]))/2})`)
  })

  innerChart.append("text").text("plan->").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", x(plandata[plandata.length-1].date) + x.bandwidth()/6).attr("y", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 30).attr("fill", "#FA7070").attr("font-size", 14).attr("transform", `rotate(90, ${x(plandata[plandata.length-1].date) + x.bandwidth()/6}, ${y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 30})`)

  if (target != undefined) { 
    const dates = Array.from(d3.union(plandata.map(d => d.date), data.map(d => d.date))) 
    target = target.filter(t => dates.includes(t.date))

    innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

    innerChart.append("g").selectAll().data(target).join("text").attr("text-anchor", "end").attr("alignment-baseline", "middle").text(d => `- ${d3.format("~s")(d.value)}`).attr("font-size", "12px").attr("x", innerWidth + 10).attr("y", d => y(d.value)).attr("fill", "currentColor")

    innerChart.append("text").text("Target").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", y(d3.min(target, d => d.value)) + 10).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(-90, ${innerWidth}, ${y(d3.min(target, d => d.value)) + 10})`)

  }

  if (inventorydata != undefined) {
    const y2 = d3.scaleLinear().domain([0,  inventorydata[0].inventory + inventorydata[1].inventory + 1]).rangeRound([3*innerHeight/4, 0]).nice()

    const leftInnerChart = svg.append("g").attr("transform", `translate(0, ${innerHeight/4})`)

    svg.append("line").attr("x1", 70).attr("y1", height).attr("x2", 70).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[0].inventory) + margin.bottom).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[0].inventory)).attr("fill", color(inventorydata[0].prodtype));

      leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[1].inventory)).attr("fill", color(inventorydata[1].prodtype));

      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[0].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[0].inventory/2) + margin.bottom).attr("fill", "#102C57").attr("font-size", 12)

      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[1].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y",  y2(inventorydata[1].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#102C57").attr("font-size", 12)
      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[0].inventory + inventorydata[1].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y",  y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)

      leftInnerChart.append("text").text(inventorydata[1].createdatstr).attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 60).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(90, 60, ${y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))})`)

    svg.append("text").text("Inventory").attr("text-anchor", "start").attr("x", 5).attr("y", height).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)
  } 


  return svg.node();

  function change(flag) {
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    }
  }
}

const drawPackEfficiencyChart = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.value)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.value)).attr("height", d => y(0) - y(d.value)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2").append("title").text(d => d.value)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d.value > 15000 ? d3.format(",.0f")(d.value) : "").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.value)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.value)})`)
      
svg.append("text").text("Sản lượng($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).value / w.workhr / 192.8 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  svg.append("text").text("Demand: 192.8 $/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)      

  if (workinghrs.length != 0) {
    const lastW =  workinghrs[workinghrs.length-1]
    innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)
  }
  
  }

  return svg.node();
}

const drawPanelcncChart1 = (data) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.left;

  const fx = d3.scaleBand().domain(new Set(data.map(d => d.date))).rangeRound([margin.left, innerWidth]).paddingInner(0.15);

  const machines = new Set(data.map(d => d.machine))

  const x = d3.scaleBand().domain(machines).rangeRound([0, fx.bandwidth()]).paddingInner(0.05);

  const color = d3.scaleOrdinal()
    .domain(machines)
    .range(d3.schemeTableau10)
    .unknown("#ccc");

  const y = d3.scaleLinear()
    .domain([0, d3.max(data, d => d.qty)])
    .rangeRound([innerHeight, 0])
    .nice();

  const svg = d3.creates("svg").attr("width", width).attr("height", height).attr("viewBox", [0, 0, width, height]).attr("style", "max-width: 100%; height: auto;")

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.append("g").selectAll().data(d3.group(data, d => d.date)).join("g").attr("transform", ([date]) => `translate(${fx(date)}, 0)`).selectAll().data(([, d]) => d).join("rect").attr("x", d => x(d.machine)).attr("y", d => y(d.qty)).attr("width", x.bandwidth()).attr("height", d => y(0) - y(d.qty)).attr("fill", d => color(d.machine)).append("title").text(d => d.qty)

  innerChart.append("g").selectAll().data(d3.group(data, d => d.date)).join("g").attr("transform", ([date]) => `translate(${fx(date)}, 0)`).selectAll().data(([, d]) => d).join("text").text(d => x.bandwidth() >= 20 ? d.qty : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.machine) + x.bandwidth()/2).attr("y", d => y(d.qty) + 8).attr("fill", "white").attr("font-size", "14px")

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(fx).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"));

  innerChart.append("g").attr("transform", `translate(${margin.left}, 0)`).call(d3.axisLeft(y).ticks(null, "s")).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", 14)).call(g => g.selectAll(".tick line").clone().attr("x2", innerWidth).attr("stroke-opacity", 0.1))

  svg.append("text").text("(sheet)").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 60).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 14)

  svg.append("text").text("rover c").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 890).attr("y", 90).attr("dy", "0.35em").attr("fill", color("rover c")).attr("font-size", 16).attr("font-weight", 600).attr("transform", d => `rotate(-90, 890, 90)`)

  svg.append("text").text("panel saw 3").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 876).attr("y", 90).attr("dy", "0.35em").attr("fill", color("panel saw 3")).attr("font-size", 16).attr("font-weight", 600).attr("transform", d => `rotate(-90, 876, 90)`)

  svg.append("text").text("panel saw 2").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 862).attr("y", 90).attr("dy", "0.35em").attr("fill", color("panel saw 2")).attr("font-size", 14).attr("font-weight", 600).attr("transform", d => `rotate(-90, 862, 90)`)

  svg.append("text").text("panel saw 1").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 848).attr("y", 90).attr("dy", "0.35em").attr("fill", color("panel saw 1")).attr("font-size", 14).attr("font-weight", 600).attr("transform", d => `rotate(-90, 848, 90)`)

  svg.append("text").text("nesting 2").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 834).attr("y", 90).attr("dy", "0.35em").attr("fill", color("nesting 2")).attr("font-size", 14).attr("font-weight", 600).attr("transform", d => `rotate(-90, 834, 90)`)

  svg.append("text").text("nesting 1").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 820).attr("y", 90).attr("dy", "0.35em").attr("fill", color("nesting 1")).attr("font-size", 14).attr("font-weight", 600).attr("transform", d => `rotate(-90, 820, 90)`)

  return svg.node();
}

const drawPanelcncChart = (data) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const x = d3.scaleBand()
    .domain(data.map(d => d.date))
    .range([0, innerWidth])
    .padding(0.1);

  const xAxis = d3.axisBottom(x).tickSizeOuter(0);

  const y = d3.scaleLinear()
    .domain([0, d3.max(data, d => d.qty)])
    .range([innerHeight, 0])
    .nice();

  const svg = d3.creates("svg").attr("width", width).attr("height", height).attr("viewBox", [0, 0, width, height]).attr("style", "max-width: 100%; height: auto;")

  const innerChart = svg.append("g").attr("class", "bars").attr("fill", "#DFC6A2").attr("transform", `translate(${margin.left}, ${margin.top})`)
  
  innerChart.append("g").selectAll("rect").data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("width", d => x.bandwidth()).attr("height", d => y(0) - y(d.qty)).append("title").text(d => d.qty)

  innerChart.append("g").attr("class", "label").attr("font-family", "sans-serif").selectAll("text").data(data).join("text").text(d => d.qty).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.qty) - 12).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600)

  innerChart.append("g").attr("class", "x-axis").attr("transform", `translate(0, ${innerHeight})`).call(xAxis).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"));
  
  svg.append("text").text("(sheet)").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 30).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 16)

  return svg.node();
}

const drawPanelcncChart2 = (data, target) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const xAxis = d3.axisBottom(x).tickSizeOuter(0);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(data, d => d.qty), d3.max(target, d => d.value)])]).range([innerHeight, 0]).nice();

  const svg = d3.creates("svg").attr("width", width).attr("height", height).attr("viewBox", [0, 0, width, height]).attr("style", "max-width: 100%; height: auto;")

  const innerChart = svg.append("g").attr("class", "bars").attr("fill", "#DFC6A2").attr("transform", `translate(${margin.left}, ${margin.top})`)
  
  innerChart.append("g").selectAll("rect").data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("width", d => x.bandwidth()).attr("height", d => y(0) - y(d.qty)).append("title").text(d => d.qty)

  innerChart.append("g").attr("class", "label").attr("font-family", "sans-serif").selectAll("text").data(data).join("text").text(d => d.qty).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.qty) - 12).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600)

  innerChart.append("g").attr("class", "x-axis").attr("transform", `translate(0, ${innerHeight})`).call(xAxis).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"));
  
  const dates = data.map(d => d.date)
  target = target.filter(t => dates.includes(t.date))

innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {if (i == target.length-1) return d.value;else {if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 15) return d.value;}}).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)

  svg.append("text").text("(sheet)").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 30).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 16)

  return svg.node();
}

const drawPanelcncEfficiecyChart = (data, manhr) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("height", d => y(0) - y(d.qty)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2").append("title").text(d => d.qty)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d.qty >= 30 ? d3.format(".0f")(d.qty) : "").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.qty)})`)
      
svg.append("text").text("Sản lượng(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).qty / w.workhr / 5 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  if (workinghrs.length != 0) {
    const lastW = workinghrs[workinghrs.length-1]
    innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)
  }
  
  svg.append("text").text("Demand: 5 sheets/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

const drawProdTypeChart = (data, rawData) => {
  const width = 300;
  const height = 200;
  const radius = 200/2;
  
  const MTDTotal = data.reduce((total, d) => total + d.value, 0)
  dates = new Set(rawData.map(d => d.date))

  const arc = d3.arc().startAngle(d => d.startAngle).endAngle(d => d.endAngle).innerRadius(radius * 0.6).outerRadius(radius - 1).padAngle(0.02).cornerRadius(3);

  const pie = d3.pie().padAngle(1/radius).sort(null).value(d => d.value);

  const color = d3.scaleOrdinal().domain(data.map(d => d.name)).range(d3.schemeTableau10)

  const svg = d3.creates("svg").attr("viewBox", [-width/2, -height/2, width, height])

  svg.append("g").selectAll().data(pie(data)).join("path").attr("fill", d => color(d.data.name)).attr("opacity", 0.8).attr("d", arc).append("title").text(d => `${d.data.name} - ${d.data.value}`);
  
  svg.append("g").attr("font-family", "sans-serif").attr("font-size", 8).attr("text-anchor", "middle").selectAll().data(pie(data)).join("text").attr("transform", d => `translate(${arc.centroid(d)})`).call(text => text.filter(d => (d.endAngle - d.startAngle) > 0.25).append("tspan").text(d => d.data.name).attr("y", "-0.5em").style("text-transform", "uppercase").attr("fill", "#f6fafc").attr("font-weight", 500)).call(text => text.filter(d => (d.endAngle - d.startAngle) > 0.25).append("tspan").attr("x", 0).attr("y", "0.8em").attr("fill-opacity", 1).attr("fill", "#f6fafc").text(d => d3.format(".3s")(d.data.value)).append("tspan").attr("x", 0).attr("y", "2em").attr("fill-opacity", 1).attr("fill", "#f6fafc").text(d => `${d3.format(".3s")(d.data.value/MTDTotal*100)}%`))

  svg.append("text").attr("text-anchor", "middle").attr("dominant-baseline", "middle").append("tspan").text(d3.format("$,.0f")(MTDTotal)).attr("y", "0.2em").attr("font-size", "20px").attr("font-weight", 500).append("tspan").text("Value Total").attr("x", "-0.25em").attr("y", "-1.5em").attr("font-size", "14px").append("tspan").text(`from ${rawData[rawData.length-1].date}`).attr("x", "-1.5em").attr("y", "2.5em").attr("font-size", "10px").attr("opacity", 0.5).append("tspan").text(`to ${rawData[0].date}`).attr("x", "1em").attr("y", "4em").attr("font-size", "10px")

  svg.append("text").text(`Working days: ${dates.size}`).attr("text-anchor", "start").attr("dominant-baseline", "middle").attr("x", -width/2).attr("y", -height/2 + 5).attr("fill", "#75485E").attr("font-size", "10px")
  
  svg.append("text").text(`Avg: ${d3.format("$,.0f")(MTDTotal/dates.size)}`).attr("text-anchor", "start").attr("dominant-baseline", "middle").attr("x", -width/2).attr("y", -height/2 + 20).attr("fill", "#75485E").attr("font-size", "10px")

  svg.append("text").text("Latest update at").attr("text-anchor", "end").attr("dominant-baseline", "middle").attr("x", width/2 - 10).attr("y", -height/2 + 5).attr("fill", "#75485E").attr("font-size", "10px")

  const latestDate = rawData[0].createdat.slice(0, 10)
  const latestTime = rawData[0].createdat.slice(11, 16)
  svg.append("text").text(latestDate).attr("text-anchor", "end").attr("dominant-baseline", "middle").attr("x", width/2 - 10).attr("y", -height/2 + 20).attr("fill", "#75485E").attr("font-size", "10px")

  svg.append("text").text(latestTime).attr("text-anchor", "end").attr("dominant-baseline", "middle").attr("x", width/2 - 10).attr("y", -height/2 + 35).attr("fill", "#75485E").attr("font-size", "10px")
    
  return svg.node();
}

const drawProductionChart = (data) => {
  const margin = {top: 30, right: 50, bottom: 10, left: 85};
  const width = 900;
  const height = 350;
  
  const barStep = 27;
  const barPadding = 3 / barStep;
  const duration = 300;
  
  const x = d3.scaleLinear().range([margin.left, width - margin.right]);
  
  const xAxis = g => g.attr("class", "x-axis").attr("transform", `translate(0,${margin.top})`).call(d3.axisTop(x).ticks(width / 80, "s")).call(g => (g.selection ? g.selection() : g).select(".domain").remove())
  
  const yAxis = g => g.attr("class", "y-axis").attr("transform", `translate(${margin.left + 0.5},0)`)
    
  const color = d3.scaleOrdinal([true, false], ["steelblue", "#aaa"]);

  const root = d3.hierarchy(data).sum(d => d.value).eachAfter(d => d.index = d.parent ? d.parent.index = d.parent.index + 1 || 0 : 0);
  
  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height]).attr("style", "max-width: 100%; height: auto;");
  
  x.domain([0, root.value]);

  svg.append("rect").attr("class", "background").attr("fill", "none").attr("pointer-events", "all").attr("width", width).attr("height", height).attr("cursor", "pointer").on("click", (event, d) => up(svg, d));

  svg.append("g").call(xAxis);

  svg.append("g").call(yAxis);

  down(svg, root);

  return svg.node();

  function down(svg, d) {
    if (!d.children || d3.active(svg.node())) return;
  
    svg.select(".background").datum(d);
  
    const transition1 = svg.transition().duration(duration);
    const transition2 = transition1.transition();
  
    const exit = svg.selectAll(".enter")
        .attr("class", "exit");

    exit.selectAll("rect")
        .attr("fill-opacity", p => p === d ? 0 : null);

    exit.transition(transition1)
        .attr("fill-opacity", 0)
        .remove();
  
    const enter = bar(svg, down, d, ".y-axis")
        .attr("fill-opacity", 0);

    enter.transition(transition1)
        .attr("fill-opacity", 1);
  
    enter.selectAll("g")
        .attr("transform", stack(d.index))
      .transition(transition1)
        .attr("transform", stagger());
  
    x.domain([0, d3.max(d.children, d => d.value)]);
  
    svg.selectAll(".x-axis").transition(transition2)
        .call(xAxis);

    enter.selectAll("g").transition(transition2)
        .attr("transform", (d, i) => `translate(0,${barStep * i})`);

    enter.selectAll("rect")
        .attr("fill", color(true))
        .attr("fill-opacity", 1)
      .transition(transition2)
        .attr("fill", d => color(!!d.children))
        .attr("width", d => x(d.value) - x(0));

  }

  function up(svg, d) {
    if (!d.parent || !svg.selectAll(".exit").empty()) return;
  
    svg.select(".background").datum(d.parent);
  
    const transition1 = svg.transition().duration(duration);
    const transition2 = transition1.transition();
  
    const exit = svg.selectAll(".enter").attr("class", "exit");
  
    x.domain([0, d3.max(d.parent.children, d => d.value)]);
  
    svg.selectAll(".x-axis").transition(transition1).call(xAxis);
  
    exit.selectAll("g").transition(transition1)
        .attr("transform", stagger());

    exit.selectAll("g").transition(transition2)
        .attr("transform", stack(d.index));
  
    exit.selectAll("rect").transition(transition1)
        .attr("width", d => x(d.value) - x(0))
        .attr("fill", color(true));
  
    exit.transition(transition2)
        .attr("fill-opacity", 0)
        .remove();
  
    const enter = bar(svg, down, d.parent, ".exit")
        .attr("fill-opacity", 0);
  
    enter.selectAll("g")
        .attr("transform", (d, i) => `translate(0,${barStep * i})`);
  
    enter.transition(transition2)
        .attr("fill-opacity", 1);
  
    enter.selectAll("rect")
        .attr("fill", d => color(!!d.children))
        .attr("fill-opacity", p => p === d ? 0 : null)
      .transition(transition2)
        .attr("width", d => x(d.value) - x(0))
        .on("end", function(p) { d3.select(this).attr("fill-opacity", 1); });
  }

  function bar(svg, down, d, selector) {
    x.domain([0, d3.max(d.children, d => d.value)]);

    const g = svg.insert("g", selector)
        .attr("class", "enter")
        .attr("transform", `translate(0,${margin.top + barStep * barPadding})`)
        .attr("text-anchor", "end")
        .style("font", "12px sans-serif");
        
    const bar = g.selectAll("g")
      .data(d.children)
      .join("g")
        .attr("cursor", d => !d.children ? null : "pointer")
        .on("click", (event, d) => down(svg, d));

    bar.append("text")
      .attr("text-anchor", "start")
      .attr("alignment-baseline", "middle")
      .attr("x", d => x(d.value) + 5)
      .attr("y", barStep * (1 - barPadding) / 2)
      .attr("font-size", "12px")
      .attr("fill","#75485E")
      .text(d => `${d3.format(",.0f")(d.value)}`);

    bar.append("text")
        .attr("x", margin.left - 6)
        .attr("y", barStep * (1 - barPadding) / 2)
        .attr("dy", ".35em")
        .attr("font-size", "12px")
        .text(d => d.data.name);

    bar.append("rect")
        .attr("x", x(0))
        .attr("width", d => x(d.value) - x(0))
        .attr("height", barStep * (1 - barPadding));
  
    return g;
  }

  function stack(i) {
    let value = 0;
    return d => {
      const t = `translate(${x(value) - x(0)},${barStep * i})`;
      value += d.value;
      return t;
    };
  }

  function stagger() {
    let value = 0;
    return (d, i) => {
      const t = `translate(${x(value) - x(0)},${barStep * i})`;
      value += d.value;
      return t;
    };
  }
}

const drawProdMtdChart = (data) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 30, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  let ar = data[data.length-1].dat
  let re = 0;
  let tr = 0;
  let th = [];
  if (ar.length > 1) {
    re = ar[ar.length-2].days 
    tr = ar[ar.length-2].value / re
    th = [{days: re + 1, value: ar[ar.length-2].value + tr}]
    for (let i = re+2; i < 27; i++) { 
      th.push({days: i, value: th[th.length-1].value + tr})
    }
  }
  const x = d3.scaleLinear().domain([1, 31]).range([0, innerWidth])
  const y = d3.scaleLinear().domain([0, d3.max(data.map(d => d.dat[d.dat.length-1]), d => d.value)]).range([innerHeight, 0]).nice();
  const color = d3.scaleOrdinal().domain(data.map(d => d.month)).range(["#FEEFAD", "#FDDE55", "#68D2E8", "#03AED2"])
  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height]);
  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)
  const area = d3.area().x(d => x(d.days)).y0(d => y(0)).y1(d => y(d.value)).curve(d3.curveCatmullRom)
  data.forEach((serie, i) => {
    innerChart.append("path").attr("d", area(serie.dat)).attr("fill", color(serie.month)).attr("fill-opacity", serie == data[data.length-1] ? 0.9 : 0.4)
    innerChart.append("text").text(`${serie.month} - $ ${d3.format(",.0f")(serie.dat[serie.dat.length-1].value)}`).attr("font-size", "14px").attr("x", (i == data.length-1 && serie.dat[serie.dat.length-1].days > 6 && serie.dat[serie.dat.length-1].days < 25) ? x(serie.dat[serie.dat.length-1].days) - 150 : x(serie.dat[serie.dat.length-1].days) + 14 ).attr("y", y(serie.dat[serie.dat.length-1].value) - 13).attr("fill", "#75485E")
    innerChart.append("line").attr("x1", x(serie.dat[serie.dat.length-1].days)).attr("y1", y(serie.dat[serie.dat.length-1].value) + 1).attr("x2", (i == data.length-1 && serie.dat[serie.dat.length-1].days > 6 && serie.dat[serie.dat.length-1].days < 25) ? x(serie.dat[serie.dat.length-1].days) - 13 : x(serie.dat[serie.dat.length-1].days) + 13).attr("y2", y(serie.dat[serie.dat.length-1].value) - 11).attr("stroke", "#75485E").attr("stroke-width", 1);
  })
  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll("text").attr("font-size", "14px")).call(g => g.append("text").text("days").attr("text-anchor", "start").attr("x", innerWidth - 10).attr("y", 16).attr("fill", "#75485E").attr("font-size", "12px").attr("font-family", "Roboto, sans-serif"))
  innerChart.append("g").call(d3.axisLeft(y).ticks(null, "s")).call(g => g.selectAll(".domain").remove()).call(g => g.append("text").text("MTD Value").attr("text-anchor", "start").attr("x", -30).attr("y", -10).attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 500).attr("font-family", "Roboto, sans-serif"))

  if (th.length > 1) {
    innerChart.append("path").attr("d", area(th)).attr("fill", color(data[data.length-1].month)).attr("fill-opacity", 0.05)
  innerChart.append("text").text(`Estimate: $ ${d3.format(",.0f")(th[th.length-1].value)}`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("x", x(th[th.length-1].days) - 20).attr("y", y(th[th.length-1].value)).attr("fill", "#75485E")
  innerChart.append("line").attr("x1",  x(th[th.length-1].days)).attr("y1", y(th[th.length-1].value) - 1).attr("x2",  x(th[th.length-1].days) - 20).attr("y2", y(th[th.length-1].value) - 3).attr("stroke", "#75485E").attr("stroke-width", 1);
  }
  innerChart.append("text").text(`AVG of This Month up to ${re}th: $ ${d3.format(",.0f")(tr)}`).attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("x", 100).attr("y", 50).attr("fill", "#75485E")
  return svg.node()
}

const drawVOPChart = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 40, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1)

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.value)]).range([innerHeight, innerHeight/2]).nice();

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height]);

  const innerChart = svg.append("g").attr("transform", `translate(${margin.left}, ${margin.top})`)

  const xAxis = innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").text((d, i) => (d.slice(0, 2) == "01") ? d.slice(2, 6) : d.slice(0, 2)).attr("font-size", (data.length > 30) ? "10px" :"12px"))

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.value)).attr("width", x.bandwidth()).attr("height", d => innerHeight - y(d.value)).attr("fill", "steelblue").append("title").text(d => d.value)

  innerChart.selectAll().data(data).join("text").text(d => (d.value > 40000) ? d3.format(",.0f")(d.value) : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value/2)).attr("dy", "0.1em").attr("fill", "white").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/2}, ${y(d.value/2)})`)

  svg.append("text").text("Production Value ($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", innerHeight).attr("dy", "0.5em").attr("fill", "steelblue").attr("font-size", 14).attr("transform", `rotate(-90, 0, ${innerHeight})`)
    
  if (manhr != undefined) {
  const workinghrs = manhr
  
  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).value / (w.manhr / 208)
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/2, 0]).nice()

  innerChart.append("line").attr("x1", 0).attr("y1", y2(2500)).attr("x2", innerWidth).attr("y2", y2(2500)).attr("stroke", "red").attr("stroke-opacity", 0.4)
  innerChart.append("text").text("Target: 2,500").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", y2(2500)).attr("dy", "-0.5em").attr("fill", "red").attr("font-weight", 600).attr("font-size", 12)

  if (data.length > 15) {
    innerChart.append("g").attr("transform", `translate(${innerWidth}, 0)`).call(d3.axisRight(y2)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").text(d => `${d3.format("~s")(d)}`).attr("font-size", "12px")).call(g => g.selectAll(".tick line").clone(true).attr("x2", -innerWidth).attr("opacity", 0.2))
  }

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs))

  innerChart.append("g").selectAll().data(workinghrs).join("circle").attr("cx", d => x(d.date) + x.bandwidth()/2).attr("cy", d => y2(d.efficiency)).attr("r", 3).attr("fill", "#75485E").append("title").text(d => `${d3.format(",.0f")(d.efficiency)}$ with manhr(${d.manhr})` )
  
  if (data.length <= 15) {
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(",.0f")(d.efficiency)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "12px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);
        
    innerChart.append("text").text("$/FTE").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(workinghrs[workinghrs.length-1].date) + x.bandwidth()/2).attr("y", y2(workinghrs[workinghrs.length-1].efficiency)).attr("dx", "1.2em").attr("dy", "0.3em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  if (data.length > 15) {
    xAxis.call(g => g.selectAll(".tick line").clone(true).attr("y2", -innerHeight).attr("opacity", 0.1))

    innerChart.append("text").text("$/FTE").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", innerHeight/2).attr("dy", "1.5em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 12)
  }
  

  }

  return svg.node();
}

const drawQualityChart = (data) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  data.map(d => d.failedrate = Math.round(d.failedqty / d.checkedqty * 100))

  const series = d3.stack().keys(d3.union(data.map(d => d.section))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).failedrate)(d3.index(data, d => d.date, d => d.section))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(series, d => d3.max(d, d => d[1]))]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(d3.schemeSet3).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.7).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()).append("title")
  .text(d => {
    const failedqty = d.data[1].get(d.key) == undefined ? "" : d.data[1].get(d.key).failedqty
    const checkedqty = d.data[1].get(d.key) == undefined ? "" : d.data[1].get(d.key).checkedqty
    return  d[1] - d[0] < 15 ? `${d.key} - ${failedqty}/${checkedqty} (${d[1]-d[0]}%)` : ""
  })

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 - 9).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").text(d =>  d[1] - d[0] >= 15 ? d.key : "").append("tspan")
    .text(d => {
      const failedqty = d.data[1].get(d.key) == undefined ? "" : d.data[1].get(d.key).failedqty
      const checkedqty = d.data[1].get(d.key) == undefined ? "" : d.data[1].get(d.key).checkedqty
      return  d[1] - d[0] >= 15 ? `${failedqty}/${checkedqty}(${d[1]-d[0]}%)` : ""
    })
    .attr("x", d => x(d.data[0]) + x.bandwidth()/2)
    .attr("dy", "1.5em")
  })

  return svg.node();
}

const drawRawwoodChart = (importdata, selectiondata) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const x = d3.scaleBand().domain(importdata.map(d => d.date)).range([0, innerWidth]).padding(0.1)

  const y = d3.scaleLinear().domain([0, d3.max(d3.union(importdata.map(d => d.qty), d3.rollups(selectiondata, D => d3.sum(D, d => d.qty), d => d.date).map(d => d[1])))]).range([innerHeight, 0]).nice();

  const svg = d3.creates("svg").attr("viewBox", [0, 0, width, height]);

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x)).call(g => g.selectAll(".domain").remove())

  innerChart.selectAll().data(importdata).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("width", x.bandwidth()/2).attr("height", d => innerHeight - y(d.qty)).attr("fill", "#DFC6A2").append("title").text(d => d.qty)

  innerChart.selectAll().data(importdata).join("text").text(d => d3.format(".1f")(d.qty)).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "-0.3em").attr("fill", "75485E").attr("font-size", 12)
      
  svg.append("text").text("Gỗ nhập").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 10).attr("fill", "#DFC6A2").attr("font-size", 14)

  svg.append("text").text("m³").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 70).attr("fill", "#75485E").attr("font-size", 14)

  if (selectiondata != undefined) {
    const selectSeries = d3.stack().keys(d3.union(selectiondata.map(d => d.woodtone))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(selectiondata, d => d.date, d => d.woodtone))

  const color = d3.scaleOrdinal().domain(selectSeries.map(d => d.key)).range(["#A0937D", "#FAEED1", "#DFC6A2"]).unknown("#ccc");

  innerChart.selectAll().data(selectSeries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/2).append("title").text(d => d[1] - d[0])

  selectSeries.forEach(serie => {
    innerChart.append("g").selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 3*x.bandwidth()/4).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("fill", "#75485E").attr("font-size", 12).text(d => (d[1] - d[0] >= 0.1) ? `${d3.format(",.1f")(d[1]-d[0])}` : "")
  })

  svg.append("text").text("Light").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("fill", color("light")).attr("font-size", 14)

  svg.append("text").text("Dark").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 50).attr("fill", color("dark")).attr("font-size", 14)
  }
  
  return svg.node();
}

const drawReededlineChart = (data, wood25data) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.tone))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.tone))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(series, d => d3.max(d, d => d[1]))]).rangeRound([innerHeight, innerHeight/3]).nice()

  const y1 = d3.scaleLinear().domain([0, d3.max(wood25data, d => d.qty)]).rangeRound([innerHeight/3, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(["#FAEED1", "#A0937D"]).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.7).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()).append("title").text(d => d3.format(".1f")(d[1] - d[0]))

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("font-weight", 600).text(d => `Σ ${d3.format(".0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").text(d => d[1] - d[0] >= 25 ?  d3.format(".0f")(d[1]-d[0]) : "")
  })

  svg.append("text").text("Gỗ (m3) cắt cho Reeded").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 5).attr("y", height/3-margin.bottom+30).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("transform", `rotate(-90, 5, ${height/3-margin.bottom+30})`)

  svg.append("text").text("Sản lượng Reeded (m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 5).attr("y", height-margin.bottom).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("transform", `rotate(-90, 5, ${height-margin.bottom})`)

      const maxOne = series[1].find(d => d[1] == d3.max(series[1], d => d[1]))
      innerChart.append("text").text("Dark").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(maxOne.data[0])-30).attr("y", y(maxOne[1]) - 15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px")
    
      innerChart.append("line").attr("x1", x(maxOne.data[0])-10).attr("y1", y(maxOne[1]) - 5).attr("x2", x(maxOne.data[0])).attr("y2", y(maxOne[1])).attr("stroke", "#75485E").attr("stroke-width", 1);
    
      innerChart.append("text").text("Bright").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(maxOne.data[0]) + x.bandwidth() + 5).attr("y", y(maxOne[1])-15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px")

      innerChart.append("line").attr("x1", x(maxOne.data[0]) + x.bandwidth() + 15).attr("y1", y(maxOne[1]) - 5).attr("x2", x(maxOne.data[0]) + x.bandwidth() - 5).attr("y2", y(maxOne[0])+15).attr("stroke", "#75485E").attr("stroke-width", 1);
  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1.5).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y1(d.qty)).curve(d3.curveStep)(wood25data));

  innerChart.append("g").attr("font-family", "san-serif").attr("font-size", 14).attr("font-weight", 600).selectAll().data(wood25data).join("text").text(d => d3.format(",.3s")(d.qty)).attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.qty)).attr("dy", "-0.1em").attr("fill", "#75485E")

  return svg.node();
}

const drawReededlineChart1 = (data, wood25data, target) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.tone))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.tone))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), d3.max(target, d => d.value)])]).rangeRound([innerHeight, innerHeight/3]).nice()

  const y1 = d3.scaleLinear().domain([0, d3.max(wood25data, d => d.qty)]).rangeRound([innerHeight/4, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key)).range(["#FAEED1", "#A0937D"]).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.7).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth())

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 10).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("font-weight", 600).text(d => `Σ ${d3.format(".0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").text(d => d[1] - d[0] >= 30 ? d3.format(".0f")(d[1]-d[0]) : "")
  })

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1.5).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y1(d.qty)).curve(d3.curveStep)(wood25data));

  innerChart.append("g").attr("font-family", "san-serif").attr("font-size", 14).attr("font-weight", 600).selectAll().data(wood25data).join("text").text(d => d3.format(",.3s")(d.qty)).attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.qty)).attr("dy", "-0.1em").attr("fill", "#75485E")

  const dates = data.map(d => d.date)
  target = target.filter(t => dates.includes(t.date))
innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {if (i == target.length-1) return d.value;else {if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 15) return d.value;}}).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)

  svg.append("text").text("Gỗ (m3) cắt cho Reeded").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 5).attr("y", height/3-margin.bottom+30).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("transform", `rotate(-90, 5, ${height/3-margin.bottom+30})`)

  svg.append("text").text("Sản lượng Reeded (m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 5).attr("y", height-margin.bottom).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("transform", `rotate(-90, 5, ${height-margin.bottom})`)

  const maxOne = series[1].find(d => d[1] == d3.max(series[1], d => d[1]))
  innerChart.append("text").text("Dark").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(maxOne.data[0])-25).attr("y", y(maxOne[1]) - 20).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px")

  innerChart.append("line").attr("x1", x(maxOne.data[0])-10).attr("y1", y(maxOne[1]) - 10).attr("x2", x(maxOne.data[0])).attr("y2", y(maxOne[1])).attr("stroke", "#75485E").attr("stroke-width", 1);

  innerChart.append("text").text("Bright").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(maxOne.data[0]) + x.bandwidth()).attr("y", y(maxOne[1])-20).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px")

  innerChart.append("line").attr("x1", x(maxOne.data[0]) + x.bandwidth() + 15).attr("y1", y(maxOne[1]) - 8).attr("x2", x(maxOne.data[0]) + x.bandwidth() - 5).attr("y2", y(maxOne[0])+15).attr("stroke", "#75485E").attr("stroke-width", 1);

  return svg.node();
}

const drawReededlineChart2 = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("height", d => y(0) - y(d.qty)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2").append("title").text(d => d.qty)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d.qty >= 60 ? d3.format(".3s")(d.qty) : "").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.qty)})`)
      
svg.append("text").text("Sản lượng(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)

    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).qty / w.workhr / 1.6 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  if (workinghrs.length != 0) {
    const lastW = workinghrs[workinghrs.length-1]
    innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)
  }

  svg.append("text").text("Demand: 1.6 m²/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

const drawSlicingVTChart = (data, target) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 20};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.prodtype).sort())).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.prodtype))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(series, d => d3.max(d, d => d[1]))]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(series.map(d => d.key).sort()).range(["#FFCCCC", "#DFC6A2", "#A0D9DE"]).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 1).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()).append("title").text(d => d[1]-d[0])

  const dateTotal = new Set(data.map(d => d.date)).size
  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").text((d, i) => (i == 0 || i == dateTotal-1 || d.slice(0, 2) == "01") ? d : d.slice(0, 2)).attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 5).attr("fill", "#75485E").attr("font-weight", 600).text(d => `${d3.format(",.0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("fill", "#75485E").text(d => {if (d[1] - d[0] >= 30) { return d3.format(",.0f")(d[1]-d[0])}})
  })

if (target != undefined) { 
  const dates = data.map(d => d.date)
target = target.filter(t => dates.includes(t.date))
innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {if (i == target.length-1) return d.value;else {if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 15) return d.value;}}).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)}
  
const maxOne = series[1].find(d => d[1] == d3.max(series[1], d => d[1]))

  svg.append("text").text("Sản lượng (m²) theo gỗ ").attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 10).attr("y", height-margin.bottom).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 300).attr("font-size", 14).attr("transform", `rotate(-90, 10, ${height-margin.bottom})`).append("tspan").text("Fir").attr("fill", color("fir")).attr("font-weight", 600).append("tspan").text(", Reeded").attr("fill", color("reeded")).attr("font-weight", 600).append("tspan").text(" với ").attr("font-weight", 300).attr("fill", "#75485E").append("tspan").text(" Target").attr("fill", "#FA7070").attr("font-weight", 600)

  return svg.node();
}

const drawSlicingVMChart2 = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("height", d => y(0) - y(d.qty)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2")

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d3.format(".3s")(d.qty)).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.qty)})`)
      
svg.append("text").text("Sản lượng(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).qty / w.workhr / 2 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".3s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  const lastW = workinghrs[workinghrs.length-1]
  innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Demand: 2 m²/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

const drawVeneerChart1 = (data, target) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.type))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).qty)(d3.index(data, d => d.date, d => d.type))

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target ? d3.max(target, d => d.value) : 0])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(["curve", "straight", "reeded", "rework", "okoumebacker"]).range(["#BC7C7C", "#E4C087", "#F6EFBD", "#A2D2DF", "#CBD2A4"]).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth())

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - 12).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("font-weight", 600).text(d => `Σ ${d3.format(",.3s")(d[1])}`)

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 14).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/2).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2 ).attr("dy", "0.1em").attr("fill", "#75485E").attr("font-size", "14px").text(d => {
      if (d[1] - d[0] >= 50 && d.key == "rework") { return `🔧${d3.format(",.3s")(d[1]-d[0])}` }
      else if (d[1] - d[0] >= 50 && d.key == "straight") { return `⏌${d3.format(",.3s")(d[1]-d[0])}` }
      else if (d[1] - d[0] >= 50 && d.key == "curve") { return `⌒${d3.format(",.3s")(d[1]-d[0])}` }
      else if (d[1] - d[0] >= 50 && d.key == "reeded") { return `≊${d3.format(",.3s")(d[1]-d[0])}` }
      else if (d[1] - d[0] >= 50 && d.key == "okoumebacker") { return `${d3.format(",.3s")(d[1]-d[0])}` }
      else { if (d[1] - d[0] >= 50 && d.key == "reeded") {return d3.format(",.3s")(d[1]-d[0])} }
    })
  })

  if (target != undefined) {
    const dates = data.map(d => d.date)
target = target.filter(t => dates.includes(t.date))
innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(target).join("text").text((d,i) => {if (i == target.length-1) return d.value;else {if (d.value != target[i+1].value && Math.abs(data.filter(t => t.date == d.date).reduce((total, n) => total + n.qty, 0) - d.value) > 50) return d.value;}}).attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y(d.value)).attr("stroke", "#75485E").attr("font-weight", 300).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)}

  svg.append("text").text("Okoume+Backer").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", color("okoumebacker")).attr("font-weight", 600).attr("font-size", 16)

  svg.append("text").text("(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 110).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 16)

  svg.append("text").text("Rework").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", color("rework")).attr("font-weight", 600).attr("font-size", 16)

  svg.append("text").text("Reeded").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 50).attr("dy", "0.35em").attr("fill", color("reeded")).attr("font-weight", 600).attr("font-size", 16)

  svg.append("text").text("Thẳng").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 70).attr("dy", "0.35em").attr("fill", color("straight")).attr("font-weight", 600).attr("font-size", 16)

  svg.append("text").text("Cong").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 90).attr("dy", "0.35em").attr("fill", color("curve")).attr("font-weight", 600).attr("font-size", 16)

  return svg.node();
}

const drawVeneerChart2 = (data, manhr) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.qty)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.qty)).attr("height", d => y(0) - y(d.qty)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2")

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d3.format(".3s")(d.qty)).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.qty)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.qty)})`)
      
svg.append("text").text("Sản lượng(m²)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d.workhr}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).qty / w.workhr / 1.5 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  if (workinghrs.length != 0) {
    const lastW = workinghrs[workinghrs.length-1]
    innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)
  }
  svg.append("text").text("Demand: 1.5 m²/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

const drawWhiteWhoodVTPChart = (data, plandata, avgdata, inventorydata, target) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 80};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const series = d3.stack().keys(d3.union(data.map(d => d.type))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).value)(d3.index(data, d => d.date, d => d.type))

  const planseries = d3.stack().keys(d3.union(plandata.map(d => d.plantype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).plan)(d3.index(plandata, d => d.date, d => d.plantype))

  const x = d3.scaleBand().domain(d3.union(data.map(d=> d.date), plandata.map(d => d.date))).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(["brand", "rh", "white", "variance"]).range(["#DFC6A2", "#A5A0DE", "#D1D1D1", "#FFBE98"]).unknown("#ccc");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  if (avgdata > 0) {
    const avgData = series[series.length-1].slice(0, -1)
  const dataAvg = avgdata
  const minDate = avgData[d3.minIndex(avgData, d => d[1])].data[0]
  innerChart.append("line").attr("x1", 0).attr("y1", y(dataAvg)).attr("x2", innerWidth - x.bandwidth() - 10).attr("y2", y(dataAvg)).attr("stroke", "#257180").attr("fill", "none").attr("stroke-opacity", 0.7)
  innerChart.append("text").text(`AVG: ${d3.format(",.0f")(dataAvg)}`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", x(minDate) + x.bandwidth()).attr("y", y(dataAvg)).attr("dy", "-0.5em").attr("fill", "#257180").attr("font-weight", 600).attr("font-size", 12)
  innerChart.append("text").text("* của cột tiền thực tế từ đầu tháng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(minDate) + 5).attr("y", y(dataAvg)).attr("dy", "-1.8em").attr("fill", "#257180").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 40000)
  }

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0]) + x.bandwidth()/3).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", 2*x.bandwidth()/3).append("title").text(d => d[1] - d[0])

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 11).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[1]) - 10).attr("dy", "0.4em") .attr("fill", "#75485E").attr("font-weight", 600).text(d => `${d3.format(",.0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", "#75485E").text(d => (d[1] - d[0] >= 10000) ? `${d3.format(",.0f")(d[1]-d[0])}` : "")
  })

  svg.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", color("rh")).attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 25).attr("dy", "0.35em").attr("fill", color("brand")).attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("White").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 45).attr("dy", "0.35em").attr("fill", color("white")).attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("Variance").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 65).attr("dy", "0.35em").attr("fill", color("variance")).attr("font-weight", 600).attr("font-size", 12)
  
  svg.append("text").text("($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 85).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12)

  svg.append("text").text("* Rà chuột vào cột để hiện value cho loại hàng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 40).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  innerChart.selectAll().data(planseries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d => y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/3).attr("stroke", "#FF9874").append("title").text(d => d[1] - d[0])

    const diffs = d3.rollups(plandata, D => { return {"current": D[0].plan + D[1].plan, "prev": D[0].plan + D[1].plan + D[0].change + D[1].change}} ,d => d.date)
    innerChart.selectAll().data(diffs).join("rect").attr("x", d => x(d[0])).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) : y(d[1].prev)).attr("height", d => y(0) - y(Math.abs(d[1].current-d[1].prev))).attr("width", x.bandwidth()/3).attr("fill", "url(#diffpattern)").attr("fill-opacity", 0.3).append("title").text(d => Math.abs(d[1].current-d[1].prev))

    innerChart.selectAll().data(diffs).join("text").text(d => {if (d[1].current > d[1].prev) {return "︽"}if (d[1].current < d[1].prev) {return "︾"}}).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d[0]) + x.bandwidth()/6).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) + 10 : y(d[1].prev) + 10).attr("font-weight", 900).attr("fill", d => {if (d[1].current > d[1].prev) {return "#3572EF"}if (d[1].current < d[1].prev) {return "#C80036"}})

  planseries.forEach(planserie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(planserie).join("text").text(d => ((d[1]-d[0]) >= 10000) ? `${d3.format(",.0f")(d[1]-d[0])}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/6).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", "#102C57").attr("transform", d => `rotate(-90, ${x(d.data[0]) + x.bandwidth()/6}, ${y(d[1]) - (y(d[1]) - y(d[0]))/2})`)
  })

  innerChart.append("text").text("plan").attr("text-anchor", "middle").attr("alignment-baseline", "start").attr("x", x(plandata[plandata.length-1].date)).attr("y", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) + 40).attr("dy", "-0.2em").attr("fill", "#FA7070").attr("font-size", 14).attr("transform", `rotate(-80, ${x(plandata[plandata.length-1].date)}, ${y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) + 40})`)

    innerChart.append("line").attr("x1", x(plandata[plandata.length-1].date)).attr("y1", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan)).attr("x2", x(plandata[plandata.length-1].date) - 5).attr("y2", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) + 28).attr("stroke", "#75485E").attr("fill", "none").attr("stroke-opacity", 0.5)

  if (target != undefined) { 
    const dates = Array.from(d3.union(plandata.map(d => d.date), data.map(d => d.date))) 
    target = target.filter(t => dates.includes(t.date))

    innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

    innerChart.append("g").selectAll().data(target).join("text").attr("text-anchor", "end").attr("alignment-baseline", "middle").text(d => `- ${d3.format("~s")(d.value)}`).attr("font-size", "12px").attr("x", innerWidth + 10).attr("y", d => y(d.value)).attr("fill", "currentColor")

    innerChart.append("text").text("Target").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", y(d3.min(target, d => d.value)) + 10).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(-90, ${innerWidth}, ${y(d3.min(target, d => d.value)) + 10})`)

  }

  if (inventorydata != undefined) {
    const y2 = d3.scaleLinear().domain([0,  inventorydata[0].inventory + inventorydata[1].inventory + 1]).rangeRound([3*innerHeight/4, 0]).nice()

    const leftInnerChart = svg.append("g")
      .attr("transform", `translate(0, ${innerHeight/4})`)

    svg.append("line").attr("x1", 70).attr("y1", height).attr("x2", 70).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[0].inventory) + margin.bottom).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[0].inventory)).attr("fill", color(inventorydata[0].prodtype));

      leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[1].inventory)).attr("fill", color(inventorydata[1].prodtype));

      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[0].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y",  y2(inventorydata[0].inventory/2) + margin.bottom).attr("fill", "#102C57").attr("font-size", 12)

      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[1].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y",  y2(inventorydata[1].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#102C57").attr("font-size", 12)

      leftInnerChart.append("text").text(`${d3.format(",.0f")(inventorydata[0].inventory + inventorydata[1].inventory)}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y",  y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)

      leftInnerChart.append("text").text(inventorydata[1].createdatstr).attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 60).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(90, 60, ${y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))})`)

    svg.append("text").text("Inventory").attr("text-anchor", "start").attr("x", 5).attr("y", height).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)
  } 

  return svg.node();

  function change(flag) {
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    }
  }
}

const drawWoodFinishVTPChart = (data, plandata, inventorydata, target) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 80};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;
  
  let flag = true;
  
  data.sort((a, b) => a.type.localeCompare(b.type))
 
  const series = d3.stack().keys(d3.union(data.map(d => d.type))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).value)(d3.index(data, d => d.date, d => d.type))

  const planseries = d3.stack().keys(d3.union(plandata.map(d => d.plantype))).value(([, D], key) => D.get(key) === undefined ? 0 : D.get(key).plan)(d3.index(plandata, d => d.date, d => d.plantype))

  const x = d3.scaleBand().domain(plandata.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0,  d3.max([d3.max(series, d => d3.max(d, d => d[1])), target == undefined ? 0 : d3.max(target, d => d.value)])]).rangeRound([innerHeight, 0]).nice()

  const color = d3.scaleOrdinal().domain(["X1-brand", "X1-rh", "X2-brand", "X2-rh", "brand", "rh"]).range(["#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE", "#DFC6A2", "#A5A0DE"]).unknown("white");

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(series).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0]) + x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - 5 : y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", 2*x.bandwidth()/3).on("mouseover", e => {flag = !flag;change(flag);}).append("title").text(d => d[1] - d[0])

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 11).selectAll().data(series[series.length-1]).join("text").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[1]) - 15).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).text(d => `${d3.format(",.0f")(d[1])}` )

  series.forEach(serie => {
    innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(serie).join("text").attr("class", d => d.key.substring(0,2)).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => d.key.startsWith("X2") ? y(d[1]) - (y(d[1]) - y(d[0]))/2 - 5 : y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", d => d.key.startsWith("X1") ? "#921A40" : "#102C57").text(d => {if (d[1] - d[0] >= 3500) { return `${d3.format(",.0f")(d[1]-d[0])}` }})
  })

  const x1data = series.filter(s => s.key.startsWith("X1"))
  const x1rhdata = x1data[x1data.length-1]
  if (x1rhdata != undefined) {
    innerChart.append("g").selectAll().data(x1rhdata).join("text").text(d => d[1] > 3500 ? `Σ${d3.format(",.0f")(d[1])}` : "").attr("class", "x1total").attr("text-anchor", "middle").attr("dominant-baseline", "hanging").attr("x", d => x(d.data[0]) + 2*x.bandwidth()/3).attr("y", d => y(d[0])).attr("dy", "0.1em").attr("fill", "#921A40").attr("opacity", 0).attr("font-size", 12)

      if (flag) {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
      } else {
        innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
        innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
      }
  }

  svg.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#A5A0DE").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 20).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 12)
     
  svg.append("text").text("Xưởng 2").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 35).attr("dy", "0.35em").attr("fill", "#102C57").attr("font-weight", 600).attr("font-size", 12) 

  svg.append("text").text("Xưởng 7").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 50).attr("dy", "0.35em").attr("fill", "#921A40").attr("font-weight", 600).attr("font-size", 12)
      
  svg.append("text").text("($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 65).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12)

  svg.append("text").text(`Total: `).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("x", 80).attr("y", 8).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 14).append("tspan").text(`$${d3.format(",.0f")(d3.sum(data, d => d.value))}`).attr("text-anchor", "start").attr("alignment-baseline", "start").attr("fill", "#75485E").attr("font-size", 16).attr("font-weight", 600)

  svg.append("text").text(" <-- Giá trị total actual dựa theo số ngày trên chart").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 200).attr("y", 8).attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  svg.append("text").text("* Rà chuột vào cột để hiện value cho loại hàng").attr("class", "disappear").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 80).attr("y", 25).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", 12).style("transition", "opacity 2s ease-out")
  setTimeout(() => d3.selectAll(".disappear").attr("opacity", 0), 5000)

  innerChart.selectAll().data(planseries).join("g").attr("fill", d => color(d.key)).attr("fill-opacity", 0.9).selectAll("rect").data(D => D.map(d => (d.key = D.key, d))).join("rect").attr("x", d => x(d.data[0])).attr("y", d =>  y(d[1])).attr("height", d => y(d[0]) - y(d[1])).attr("width", x.bandwidth()/3).attr("stroke", "#FF9874").on("mouseover", e => {flag = !flag;change(flag);})

    const diffs = d3.rollups(plandata, D => { return {"current": D[0].plan + D[1].plan, "prev": D[0].plan + D[1].plan + D[0].change + D[1].change}} ,d => d.date)
    innerChart.selectAll().data(diffs).join("rect").attr("x", d => x(d[0])).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) : y(d[1].prev)).attr("height", d => y(0) - y(Math.abs(d[1].current-d[1].prev))).attr("width", x.bandwidth()/3).attr("fill", "url(#diffpattern)").attr("fill-opacity", 0.6).append("title").text(d => Math.abs(d[1].current-d[1].prev))
    
    innerChart.selectAll().data(diffs).join("text").text(d => {if (d[1].current > d[1].prev) {return "︽"}if (d[1].current < d[1].prev) {return "︾"}}).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d[0]) + x.bandwidth()/6).attr("y", d => d[1].current >= d[1].prev ? y(d[1].current) + 10 : y(d[1].prev) + 10).attr("font-weight", 900).attr("fill", d => {if (d[1].current > d[1].prev) {return "#3572EF"}if (d[1].current < d[1].prev) {return "#C80036"}})
        
    planseries.forEach(planserie => {
      innerChart.append("g").attr("font-family", "sans-serif").attr("font-size", 12).selectAll().data(planserie).join("text").text(d => `${d3.format(",.0f")(d[1]-d[0])}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", d => x(d.data[0]) + x.bandwidth()/6).attr("y", d => y(d[1]) - (y(d[1]) - y(d[0]))/2).attr("dy", "0.1em").attr("fill", "#102C57").attr("transform", d => `rotate(-90, ${x(d.data[0]) + x.bandwidth()/6}, ${y(d[1]) - (y(d[1]) - y(d[0]))/2})`)
  })

  innerChart.append("text").text("plan -->").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", x(plandata[plandata.length-1].date) + x.bandwidth()/6).attr("y", y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 30).attr("fill", "#FA7070").attr("font-size", 14).attr("transform", `rotate(90, ${x(plandata[plandata.length-1].date) + x.bandwidth()/6}, ${y(plandata[plandata.length-1].plan + plandata[plandata.length-2].plan) - 30})`)

  if (target != undefined) { 
    const dates = Array.from(d3.union(plandata.map(d => d.date), data.map(d => d.date))) 
    target = target.filter(t => dates.includes(t.date))

    innerChart.selectAll().data(target).join("line").attr("x1", d => x(d.date)).attr("y1", d => y(d.value)).attr("x2", d => x(d.date) + x.bandwidth()).attr("y2", d => y(d.value)).attr("stroke", "#FA7070").attr("fill", "none").attr("stroke-opacity", 0.5)

    innerChart.append("g").selectAll().data(target).join("text").attr("text-anchor", "end").attr("alignment-baseline", "middle").text(d => `- ${d3.format("~s")(d.value)}`).attr("font-size", "12px").attr("x", innerWidth + 10).attr("y", d => y(d.value)).attr("fill", "currentColor")

    innerChart.append("text").text("Target").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", innerWidth).attr("y", y(d3.min(target, d => d.value)) + 10).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(-90, ${innerWidth}, ${y(d3.min(target, d => d.value)) + 10})`)
  }
  
  if (inventorydata != undefined) {

    const y2 = d3.scaleLinear().domain([0, d3.sum(inventorydata, d => d.inventory)]).rangeRound([3*innerHeight/4, 0]).nice()

    const leftInnerChart = svg.append("g")
      .attr("transform", `translate(0, ${innerHeight/4})`)

    svg.append("line").attr("x1", 70).attr("y1", height).attr("x2", 70).attr("y2", 0).attr("stroke", "black").attr("stroke-opacity", 0.2)

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[0].inventory) + margin.bottom).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[0].inventory)).attr("fill", color(inventorydata[0].type));

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[1].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[1].inventory)).attr("fill", color(inventorydata[1].type));

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[2].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory)) - 5).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[2].inventory)).attr("fill", color(inventorydata[2].type));

    leftInnerChart.append("rect").attr("x", 10).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("width", 45).attr("height", 3*innerHeight/4 - y2(inventorydata[3].inventory)).attr("fill", color(inventorydata[3].type));

    leftInnerChart.append("text").text((inventorydata[0].inventory != 0) ? `${d3.format(",.0f")(inventorydata[0].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[0].inventory/2) + margin.bottom).attr("fill", "#921A40").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[1].inventory != 0) ? `${d3.format(",.0f")(inventorydata[1].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[1].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory))).attr("fill", "#921A40").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[2].inventory != 0) ? `${d3.format(",.0f")(inventorydata[2].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[2].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory)) - 5).attr("fill", "#102C57").attr("font-size", 12)

    leftInnerChart.append("text").text((inventorydata[3].inventory != 0) ? `${d3.format(",.0f")(inventorydata[3].inventory)}` : "").attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[3].inventory/2) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("fill", "#102C57").attr("font-size", 12)

    leftInnerChart.append("text").text(`${d3.format(",.0f")(d3.sum(inventorydata, d => d.inventory))}`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("x", 32).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)

    leftInnerChart.append("text").text(inventorydata[0].createdat).attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 60).attr("y", y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5).attr("fill", "#FA7070").attr("font-size", 12).attr("transform", `rotate(90, 60, ${y2(inventorydata[3].inventory) + margin.bottom - (3*innerHeight/4 - y2(inventorydata[0].inventory + inventorydata[1].inventory + inventorydata[2].inventory)) - 5})`)
    
    svg.append("text").text("Inventory").attr("text-anchor", "start").attr("x", 5).attr("y", height).attr("dy", "-0.35em").attr("fill", "#75485E").attr("font-size", 12)
  } 


  return svg.node();

  function change(flag) {
    if (flag) {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 0))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 1))
    } else {
      innerChart.call(g => g.selectAll(".X1").attr("opacity", 1))
      innerChart.call(g => g.selectAll(".x1total").attr("opacity", 0))
    }
  }
}

const drawWfEfficiencyChart = (data, manhr) => {
  if (data == undefined) return;
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 20, bottom: 20, left: 40};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const dates = new Set(data.map(d => d.date)) 

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const y = d3.scaleLinear().domain([0, d3.max(data, d => d.value)]).rangeRound([innerHeight, innerHeight/3]).nice()

  const svg = d3.creates("svg")
    .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.selectAll().data(data).join("rect").attr("x", d => x(d.date)).attr("y", d => y(d.value)).attr("height", d => y(0) - y(d.value)).attr("width", x.bandwidth()/2).attr("fill", "#DFC6A2").append("title").text(d => d.value)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").remove()).call(g => g.selectAll("text").attr("font-size", "12px"))

  innerChart.append("g").selectAll().data(data).join("text").text(d => d.value > 15000 ? d3.format(",.0f")(d.value) : "").attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()/4).attr("y", d => y(d.value)).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "12px").attr("font-weight", 600).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()/4}, ${y(d.value)})`)
      
svg.append("text").text("Sản lượng($)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 30).attr("dy", "0.35em").attr("fill", "#DFC6A2").attr("font-weight", 600).attr("font-size", 14)
 
  if (manhr != undefined) {
    const workinghrs = manhr.filter(d => dates.has(d.date))
    const y1 = d3.scaleLinear().domain([0, d3.max(manhr, d => d.workhr)]).rangeRound([innerHeight, innerHeight/3]).nice()

    innerChart.append("g").selectAll().data(workinghrs).join("rect").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y1(d.workhr)).attr("height", d => y1(0) - y1(d.workhr)).attr("width", x.bandwidth()/2).attr("fill", "#90D26D").attr("fill-opacity", 0.3)
      
    innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `👷 ${d.hc} = ${d3.format(".0f")(d.workhr)}h`).attr("text-anchor", "end").attr("alignment-baseline", "middle").attr("x", d => x(d.date) + x.bandwidth()*3/4).attr("y", d => y1(d.workhr)).attr("fill", "#75485E").attr("font-size", 12).attr("transform", d => `rotate(-90, ${x(d.date) + x.bandwidth()*3/4 }, ${y1(d.workhr)})`)

    svg.append("text").text("manhr (h)").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 55).attr("dy", "0.35em").attr("fill","#90D26D").attr("font-weight", 600).attr("font-size", 14)

  workinghrs.forEach(w => {
    w.efficiency = data.find(d => d.date == w.date).value / w.workhr / 81.4 * 100
  })

  const y2 = d3.scaleLinear().domain(d3.extent(workinghrs, d => d.efficiency)).rangeRound([innerHeight/3, 0]).nice()

  innerChart.append("path").attr("fill", "none").attr("stroke", "#75485E").attr("stroke-width", 1).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => y2(d.efficiency)).curve(d3.curveCatmullRom)(workinghrs));

  innerChart.append("g").selectAll().data(workinghrs).join("text").text(d => `${d3.format(".2s")(d.efficiency)}%`).attr("text-anchor", "middle").attr("alignment-baseline", "middle").attr("font-size", "14px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => y2(d.efficiency)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  const lastW = workinghrs[workinghrs.length-1]
  innerChart.append("text").text("Efficiency").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(lastW.date) + x.bandwidth()/2 - 5).attr("y", y2(lastW.efficiency) - 15).attr("dy", "0.35em").attr("fill","#75485E").attr("font-weight", 600).attr("font-size", 12)

  svg.append("text").text("Demand: 81.4 $/h").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-weight", 600).attr("font-size", 14)
  }

  return svg.node();
}

const drawWoodRecoveryChart = (data) => {
  const width = 900;
  const height = 350;
  const margin = {top: 20, right: 30, bottom: 20, left: 30};
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const rhdata = data.filter(d => d.prodtype == "rh")
  const branddata = data.filter(d => d.prodtype == "brand")

  const x = d3.scaleBand().domain(data.map(d => d.date)).range([0, innerWidth]).padding(0.1);

  const yBrand = d3.scaleLinear().domain(d3.extent(branddata, d => d.rate)).range([innerHeight, innerHeight/2 + 20]).nice();

  const yRh = d3.scaleLinear().domain(d3.extent(rhdata, d => d.rate)).range([innerHeight/2 - 20, 0]).nice();

  const color = d3.scaleOrdinal().domain(data.map(d => d.prodtype)).range(["#DFC6A2", "#A5A0DE"]);

  const svg = d3.creates("svg")
      .attr("viewBox", [0, 0, width, height])

  const innerChart = svg.append("g")
    .attr("transform", `translate(${margin.left}, ${margin.top})`)

  innerChart.append("g").attr("transform", `translate(0, ${innerHeight/2 - 20})`).call(d3.axisBottom(x).tickSizeOuter(0)).call(g => g.selectAll(".domain").attr("transform", `translate(0, 13)`)).call(g => g.selectAll("text").attr("font-size", "12px").clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6)).call(g => g.selectAll(".tick line").clone().attr("transform", `translate(0, 20)`))
    
  innerChart.append("path").attr("fill", "none").attr("stroke", d => color("brand")).attr("stroke-width", 1.5).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => yBrand(d.rate)).curve(d3.curveCatmullRom)(branddata));

  innerChart.append("path").attr("fill", "none").attr("stroke", d => color("rh")).attr("stroke-width", 1.5).attr("d", d => d3.line().x(d => x(d.date) + x.bandwidth()/2).y(d => yRh(d.rate)).curve(d3.curveCatmullRom)(rhdata));

  innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(branddata).join("text").text(d => `${d.rate}%`).attr("font-size", "12px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => yBrand(d.rate)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  innerChart.append("g").attr("stroke-linecap", "round").attr("stroke-linejoin", "round").attr("text-anchor", "middle").selectAll().data(rhdata).join("text").text(d => `${d.rate}%`).attr("font-size", "12px").attr("dy", "0.35em").attr("x", d => x(d.date) + x.bandwidth()/2).attr("y", d => yRh(d.rate)).clone(true).lower().attr("fill", "none").attr("stroke", "white").attr("stroke-width", 6);

  innerChart.append("text").text("Brand").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(branddata[branddata.length-1].date) + x.bandwidth()/2 + 15).attr("y", yBrand(branddata[branddata.length-1].rate) - 15).attr("dy", "0.35em").attr("fill", color("brand")).attr("font-size", "14px").attr("font-weight", 600)

    innerChart.append("text").text("RH").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", x(rhdata[rhdata.length-1].date) + x.bandwidth()/2 + 15).attr("y", yRh(rhdata[rhdata.length-1].rate) - 15).attr("dy", "0.35em").attr("fill", color("rh")).attr("font-size", "14px").attr("font-weight", 600)

  svg.append("text").text("Target: 60%").attr("text-anchor", "start").attr("alignment-baseline", "middle").attr("x", 0).attr("y", 5).attr("dy", "0.35em").attr("fill", "#75485E").attr("font-size", "14px").attr("font-weight", 600)

  return svg.node();
}

const drawWoodRemainChart = (data) => {
  const width = 300;
  const height = 200;
  const radius = 200/2;

  const arc = d3.arc().startAngle(d => d.startAngle).endAngle(d => d.endAngle).innerRadius(radius * 0.67).outerRadius(radius - 1).padAngle(0.02).cornerRadius(3);

  const pie = d3.pie().padAngle(1/radius).sort(null).value(d => d.value);

  const color = d3.scaleOrdinal().domain(data.map(d => d.name)).range(d3.schemeTableau10)

  const svg = d3.create("svg").attr("viewBox", [-width/2, -height/2, width, height])

  svg.append("g").selectAll().data(pie(data)).join("path").attr("fill", d => color(d.data.name)).attr("d", arc).append("title").text(d => d.data.value);
  
  svg.append("g").attr("font-family", "sans-serif").attr("font-size", 8).attr("text-anchor", "middle").selectAll().data(pie(data)).join("text").attr("transform", d => `translate(${arc.centroid(d)})`).call(text => text.append("tspan").text(d => `${d.data.name}ly`).attr("y", "-0.4em").style("text-transform", "capitalize").attr("fill", "#f6fafc").attr("font-weight", 500)).call(text => text.filter(d => (d.endAngle - d.startAngle) > 0.25).append("tspan").attr("x", 0).attr("y", "0.8em").attr("fill-opacity", 0.8).attr("fill", "#f6fafc").text(d => d3.format(".3f")(d.data.value)))

  const remainTotal = d3.format(".5s")(data.reduce((total, d) => total + d.value, 0))

  svg.append("text").attr("text-anchor", "middle").attr("dominant-baseline", "middle").append("tspan")  .text(remainTotal).attr("y", "0.2em").attr("font-size", "24px").attr("font-weight", 500).append("tspan").text("m³").attr("x", "4.3em").attr("y", "-0.3em").attr("font-size", "10px").append("tspan").text("Còn lại").attr("x", "-0.25em").attr("y", "-1.5em").attr("font-size", "14px")
    

  return svg.node();
}

