function хэвлэ(тоо) {
  console.log(тоо);
}

function факториал(н) {
  if (н <= 1) {
    return 1;
  }

  const бодсон = н * факториал(н - 1);
  хэвлэ(бодсон);
  return бодсон;
}

function үндсэн() {
  return факториал(12); // 12! = 479001600
}

// Measure execution time
const start = performance.now();
const result = үндсэн();
const end = performance.now();

console.log("Final result:", result);
console.log(`Execution time: ${(end - start).toFixed(5)} milliseconds`);
