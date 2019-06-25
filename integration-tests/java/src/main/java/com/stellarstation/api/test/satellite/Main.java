package com.stellarstation.api.test.satellite;

import java.util.concurrent.CountDownLatch;

public class Main {
  public static void main(String[] args) {
    final CountDownLatch latch = new CountDownLatch(1);
    latch.countDown();
    latch.countDown();
  }

  private Main() {}
}
