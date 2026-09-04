# TANEBI 95 Native

브라우저가 아닌 x86-64 UEFI 부팅 환경에서 실행되는 TANEBI 95 네이티브 프리뷰다. `gopher-os` Stage 1의 Rust `no_std` UEFI 경계를 참고하고, TANEBI 프로그램을 빌드 과정에서 실행해 시스템 매니페스트를 만든다. 커널은 GOP에서 프레임버퍼 주소를 받은 뒤 `ExitBootServices`로 펌웨어 부팅 서비스를 종료하고 데스크톱을 직접 그린다.

## 빌드

```powershell
.\native\tanebi95\scripts\build.ps1
```

산출물:

- `build/native/esp/EFI/BOOT/BOOTX64.EFI`
- `build/native/tanebi95.img` — 32 MiB FAT16 UEFI 부팅 이미지

## 실행

QEMU와 x86-64 EDK2 펌웨어가 필요하다.

```powershell
.\native\tanebi95\scripts\run.ps1
```

키보드:

- `S`: 시작 메뉴
- `T`: TANEBI Studio 창 열기/닫기
- `Esc`: 전원 화면

## 현재 경계

이 버전은 실제로 UEFI에서 부팅한 뒤 펌웨어 부팅 서비스를 종료하는 bare-metal Stage 1이다. 이후 화면은 프레임버퍼 메모리에 직접 쓰고 키보드는 PS/2 I/O 포트에서 scan code를 읽는다. TANEBI 언어 코어는 아직 Ring 3 프로세스로 실행되지 않는다. 현재는 동일 저장소의 Go 인터프리터가 `system.tanebi`를 빌드 시 실행하고, 그 결정론적 결과를 커널에 포함한다. Ring 3 TANEBI 런타임은 페이지 테이블·syscall·프로세스 로더 이후 단계다.

Microsoft Windows나 Microsoft 에셋을 포함하지 않는 독자적인 95풍 데스크톱이다.
