package expander

// ================
// IDEA
// ================
//
// Sneaky sneaky just pipe to Deno
// and replace call site with result.
//
// INPUT
//   (macro double-sum [x y]
//     `(+ (add ,x ,y) (add ,x ,y))`)
//
// OUTPUT
//   const double_sum = (x, y) =>
//     eval(`${add(x, y)} + ${add(x, y)}`)
//
// ================
